package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/formats"
)

type tagSet map[string]bool

func (tags tagSet) Set(value string) error {
	tags[value] = true
	return nil
}

func (tags tagSet) String() string {
	if tags == nil {
		return ""
	}
	var b strings.Builder
	var sep string
	for k := range tags {
		b.WriteString(sep)
		b.WriteString(k)
		sep = ","
	}
	return b.String()
}

func (tags tagSet) Contains(tag string) bool {
	_, found := tags[tag]
	return found
}

var (
	tags     = make(tagSet)
	modeMeta = false
)

func main() {
	flag.Var(tags, "t", "EXIF tags to print")
	flag.BoolVar(&modeMeta, "m", false, "Use MetaData parser")
	flag.Parse()

	path := flag.Arg(0)
	s, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Cannot access %s: %s", path, err)
	}
	if s.IsDir() {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if modeMeta {
				in, err := os.Open(path)
				if err != nil {
					return err
				}
				defer in.Close()
				var meta domain.MediaMetaData
				f, found := domain.FormatForExt(filepath.Ext(path))
				if !found {
					// SKip
					return nil
				}
				if err := f.DecodeMetaData(in, &meta); err != nil {
					return err
				}
				fmt.Printf("%s: DateTaken=%v\n", path, meta.DateTaken)
				fmt.Printf("%s: Orientation=%v\n", path, meta.Orientation)
				fmt.Printf("%s: Location=%v\n", path, meta.Location)
			} else {
				formats.PrintExif(path, func(name, value string) {
					if len(tags) == 0 || tags.Contains(name) {
						fmt.Printf("%s: %s=%s\n", path, name, value)
					}
				})
			}
			return nil
		})
	} else {
		if modeMeta {
			in, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer in.Close()
			var meta domain.MediaMetaData
			if err := domain.JPEG.DecodeMetaData(in, &meta); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  DateTaken=%s\n", meta.DateTaken)
			fmt.Printf("  Orientation=%s\n", meta.Orientation)
			fmt.Printf("  Location=%s\n", meta.Location)
		} else {
			formats.PrintExif(path, func(name, value string) {
				if len(tags) == 0 || tags.Contains(name) {
					fmt.Printf("   %s=%s\n", name, value)
				}
			})
		}
	}
}
