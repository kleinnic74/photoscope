package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/library/boltstore"
	"bitbucket.org/kleinnic74/photos/logging"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

var (
	dbName = "photos.db"

	libDir  string
	verbose bool
	glob    string
	re      *regexp.Regexp

	logger *zap.Logger

	matcher func(string) bool
)

func init() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s  [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&libDir, "l", "gophotos", "Path to photo library")
	flag.StringVar(&glob, "g", "", "Glob file pattern, only matching files will be considered")
	var reStr string
	flag.StringVar(&reStr, "r", "", "Regexp file pattern, only matching files will be considered")
	flag.BoolVar(&verbose, "v", false, "be more verbose")
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	logger = logging.From(context.Background())

	var matchers []func(string) bool
	if len(glob) > 0 {
		_, err := filepath.Match(glob, "test")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad glob '%s': %s", glob, err)
			flag.Usage()
			os.Exit(1)
		}
		matchers = append(matchers, func(path string) bool {
			match, _ := filepath.Match(glob, path)
			return match
		})
	}
	if len(reStr) > 0 {
		re, err := regexp.Compile(reStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad regular expression '%s': %s", reStr, err)
			flag.Usage()
			os.Exit(1)
		}
		matchers = append(matchers, re.MatchString)
	}
	if len(matchers) > 0 {
		matcher = func(path string) bool {
			for _, m := range matchers {
				if m(path) {
					return true
				}
			}
			return false
		}
	} else {
		matcher = func(path string) bool {
			return true
		}
	}
}

func main() {
	dbfile := filepath.Join(libDir, dbName)
	if _, err := os.Stat(dbfile); err != nil {
		logger.Fatal("No DB file found", zap.String("db", dbfile))
	}
	db, err := bolt.Open(dbfile, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		logger.Fatal("Failed to initialize library", zap.Error(err))
	}
	defer db.Close()

	store, err := boltstore.NewBoltStore(db)
	if err != nil {
		logger.Fatal("Failed to initialize library", zap.Error(err))
	}

	var count int
	for _, dir := range flag.Args() {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if !matcher(path) {
					// Skip, does not match glob
					return nil
				}
				_, err := formatOf(path)
				if err != nil {
					if verbose {
						logger.Info("Skipping unknown format", zap.String("path", path))
					}
					return nil
				}
				in, err := os.Open(path)
				if err != nil {
					return err
				}
				defer in.Close()
				count++
				_, hash, err := library.LoadContent(in)
				if err != nil {
					return err
				}
				if id, found := store.Exists(hash); found {
					logger.Info("File in library", zap.String("path", path), zap.String("id", string(id)))
				} else {
					if verbose {
						logger.Info("File not in library", zap.String("path", path))
					}
				}
				if count%100 == 0 && count > 0 {
					logger.Info("Progress", zap.Int("files", count))
				}
			}
			return nil
		})
		if err != nil {
			logger.Fatal("Error while walking directory", zap.String("root", dir), zap.Error(err))
		}
	}
}

func formatOf(path string) (domain.Format, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	return domain.FormatOf(in)
}
