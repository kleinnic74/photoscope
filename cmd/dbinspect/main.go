package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	_ "bitbucket.org/kleinnic74/photos/domain"

	bolt "go.etcd.io/bbolt"
)

var (
	dbName = "photos.db"

	libDir string

	bucket string

	keyAcceptor = func(string) bool { return true }

	exe command

	args []string

	printKey   bool
	printValue bool
	keyFilter  string

	commands = []command{
		{"buckets", listBuckets, func() *flag.FlagSet {
			flags := flag.NewFlagSet("buckets", flag.ExitOnError)
			flags.StringVar(&bucket, "b", "", "Bucket to inspect")
			return flags
		}, true},
		{"entries", listEntries, func() *flag.FlagSet {
			cmdEntries := flag.NewFlagSet("entries", flag.ExitOnError)
			cmdEntries.StringVar(&bucket, "b", "photos", "Bucket to inspect")
			cmdEntries.BoolVar(&printKey, "k", false, "Output keys")
			cmdEntries.BoolVar(&printValue, "v", false, "Output value")
			cmdEntries.StringVar(&keyFilter, "kf", "", "Key regex filter")
			return cmdEntries
		}, true},
		{"delete", deleteEntries, func() *flag.FlagSet { return nil }, false},
		{"deleteBucket", deleteBucket, func() *flag.FlagSet { return nil }, false},
	}
)

func getCommand(args []string) (command, *flag.FlagSet, error) {
	if len(args) == 0 {
		return commands[0], commands[0].flags(), nil
	}
	for i := range commands {
		if args[0] == commands[i].name {
			return commands[i], commands[i].flags(), nil
		}
	}
	return command{}, nil, fmt.Errorf("No such command: %s", args[0])
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [command] [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "[command] is one of\n")
		for _, c := range commands {
			fmt.Fprintf(os.Stderr, "\t%s\n", c.name)
		}
		flag.PrintDefaults()
	}
	flag.StringVar(&libDir, "l", "gophotos", "Path to photo library")

	flag.Parse()

	var err error
	var flags *flag.FlagSet
	exe, flags, err = getCommand(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		flag.Usage()
		os.Exit(1)
	}
	if flag.NArg() > 0 {
		args = flag.Args()[1:]
	}
	fmt.Fprintf(os.Stderr, "executing: %s [args=%s] [exe=%v]\n", exe.name, args, exe)
	if flags != nil {
		flags.Parse(args)
		args = flags.Args()
	}
}

type cmdFunc func(*bolt.Tx) error

type flagSetFunc func() *flag.FlagSet
type command struct {
	name     string
	run      cmdFunc
	flags    flagSetFunc
	readonly bool
}

func listBuckets(tx *bolt.Tx) error {
	if bucket != "" {
		return tx.Bucket([]byte(bucket)).ForEach(func(k, v []byte) error {
			if v == nil {
				fmt.Fprintln(os.Stdout, string(k))
			}
			return nil
		})
	} else {
		return tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			fmt.Fprintln(os.Stdout, string(name))
			return nil
		})
	}
}

func deleteBucket(tx *bolt.Tx) (err error) {
	for _, b := range args {
		if bucket != "" {
			fmt.Fprintf(os.Stderr, "Deleting subbucket %s/%s", bucket, b)
			err = tx.Bucket([]byte(bucket)).DeleteBucket([]byte(b))
		} else {
			fmt.Fprintf(os.Stderr, "Deleting bucket %s", b)
			err = tx.DeleteBucket([]byte(b))
		}
		if err != nil {
			return
		}
	}
	return
}

type stats struct {
	count      int
	badKeys    int
	zeroValues int
}

func (s stats) Add(sub stats) (out stats) {
	out.badKeys = s.badKeys + sub.badKeys
	out.count = s.count + sub.count
	out.zeroValues = s.zeroValues + sub.zeroValues
	return
}

func listEntries(tx *bolt.Tx) error {
	if keyFilter != "" {
		keyRE, err := regexp.Compile(keyFilter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad key filter RE: %s\n", err)
			os.Exit(2)
		}
		keyAcceptor = keyRE.MatchString
	}

	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("No such bucket: %s", bucket)
	}
	var s stats
	defer func() {
		fmt.Printf("  %d entries\n", s.count)
		fmt.Printf("  %d bad keys\n", s.badKeys)
		fmt.Printf("  %d zero values\n", s.zeroValues)
	}()
	subStats, err := walkBucket(tx, b)
	s = s.Add(subStats)
	return err
}

func walkBucket(tx *bolt.Tx, b *bolt.Bucket) (s stats, err error) {
	sep, end := "", ""
	err = b.ForEach(func(k, v []byte) error {
		if !keyAcceptor(string(k)) {
			return nil
		}
		if v == nil {
			// This is a bucket
			subStats, err := walkBucket(tx, b.Bucket(k))
			if err != nil {
				return err
			}
			s = s.Add(subStats)
		}
		s.count++
		if len(k) == 0 {
			s.badKeys++
		}
		if len(v) == 0 {
			s.zeroValues++
		}
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
		return nil
	})
	return
}

func deleteEntries(tx *bolt.Tx) error {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("No such bucket: %s", bucket)
	}
	return b.ForEach(func(k, v []byte) error {
		if err := b.Delete(k); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete %s\n", string(k))
			return err
		}
		fmt.Fprintf(os.Stderr, "deleted %s\n", string(k))
		return nil
	})
}

func main() {
	var err error
	var db *bolt.DB
	dbPath := filepath.Join(libDir, dbName)
	if db, err = bolt.Open(dbPath, 0600, &bolt.Options{ReadOnly: exe.readonly}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open Bolt DB at %s: %s", dbPath, err)
		os.Exit(1)
	}
	defer db.Close()

	if exe.readonly {
		err = db.View(exe.run)
	} else {
		err = db.Update(exe.run)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while executing: %s", err)
		os.Exit(1)
	}
}
