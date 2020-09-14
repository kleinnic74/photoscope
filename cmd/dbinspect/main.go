package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	_ "bitbucket.org/kleinnic74/photos/domain"

	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

var (
	dbName = "photos.db"

	libDir string

	bucket string

	keyAcceptor = func(string) bool { return true }

	printKey    bool
	printValue  bool
	deleteEntry bool
	verbose     bool

	logger *zap.Logger
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s  [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&libDir, "l", "gophotos", "Path to photo library")
	flag.BoolVar(&verbose, "v", false, "List photos")
	flag.StringVar(&bucket, "b", "photos", "Bucket to inspect")
	flag.BoolVar(&printKey, "k", false, "Output keys")
	flag.BoolVar(&printValue, "V", false, "Output value")
	flag.BoolVar(&deleteEntry, "d", false, "Delete entry")
	var keyFilter string
	flag.StringVar(&keyFilter, "kf", "", "Key regex filter")

	flag.Parse()

	if keyFilter != "" {
		keyRE, err := regexp.Compile(keyFilter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad key filter RE: %s\n", err)
			os.Exit(2)
		}
		keyAcceptor = keyRE.MatchString
	}
	logger = logging.From(context.Background())
}

func main() {
	db, err := bolt.Open(filepath.Join(libDir, dbName), 0600, nil)
	if err != nil {
		logger.Fatal("Failed to initialize library", zap.Error(err))
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			if bucket != "" && string(name) != bucket {
				return nil
			}
			fmt.Fprintf(os.Stderr, "Found bucket: %s\n", string(name))
			var count int
			var badKeys int
			var zeroValues int
			err := b.ForEach(func(k, v []byte) error {
				if !keyAcceptor(string(k)) {
					return nil
				}
				count++
				if len(k) == 0 {
					badKeys++
				}
				if len(v) == 0 {
					zeroValues++
				}
				sep, end := "", ""
				if printKey {
					fmt.Fprintf(os.Stdout, "%s", string(k))
					sep = ":"
					end = "\n"
				}
				if printValue {
					fmt.Fprintf(os.Stdout, "%s%s", sep, string(v))
					end = "\n"
				}
				fmt.Fprintf(os.Stdout, "%s", end)
				if deleteEntry {
					fmt.Fprintf(os.Stderr, "deleted %s\n", string(k))
					if err := b.Delete(k); err != nil {
						return err
					}
				}
				return nil
			})
			fmt.Printf("  %d entries\n", count)
			fmt.Printf("  %d bad keys\n", badKeys)
			fmt.Printf("  %d zero values\n", zeroValues)
			return err
		})
	})
}
