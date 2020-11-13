package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/formats"
	"bitbucket.org/kleinnic74/photos/geocoding"
	"bitbucket.org/kleinnic74/photos/geocoding/openstreetmap"
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

type Action func(path string) error

func printExifData(path string) error {
	formats.PrintExif(path, func(name, value string) {
		if len(tags) == 0 || tags.Contains(name) {
			fmt.Printf("%s: %s=%s\n", path, name, value)
		}
	})
	return nil
}

func parseMetaAndPrint(resolver geocoding.Resolver) Action {
	return func(path string) error {
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
		if lookupAddress && meta.Location != nil {
			a, found, _ := resolver.ReverseGeocode(context.Background(), meta.Location.Lat, meta.Location.Long)
			if found {
				fmt.Printf("%s: Address: %v\n", path, a)
			}
		}
		return nil
	}
}

var (
	tags          = make(tagSet)
	modeMeta      = false
	lookupAddress = false
	dumpQT        = false
)

func main() {
	flag.Var(tags, "t", "EXIF tags to print")
	flag.BoolVar(&modeMeta, "m", false, "Use MetaData parser")
	flag.BoolVar(&lookupAddress, "a", false, "Resolve location to an address")
	flag.BoolVar(&dumpQT, "q", false, "Produce an SVG of the geo quadtree")
	flag.Parse()

	var action Action
	if modeMeta {
		var resolver geocoding.Resolver
		if lookupAddress {
			osm := openstreetmap.NewResolver("de", "en")
			cache := geocoding.NewGeoCache(osm)
			defer func() {
				fmt.Fprintf(os.Stderr, "  Quadtree cache hits: %d\n", cache.Hits)
				fmt.Fprintf(os.Stderr, "  Quadtree cache misses: %d\n", cache.Misses)
				fmt.Fprintf(os.Stderr, "  Quadtree performance: %f%%\n", float64(cache.Hits)/float64(cache.Hits+cache.Misses))
				if dumpQT {
					out, err := os.Create("qt.svg")
					if err == nil {
						view := NewGeoView(out)
						cache.Visit(view)
						view.Close()
					}
				}
			}()
			resolver = cache
		}
		action = parseMetaAndPrint(resolver)
	} else {
		action = printExifData
	}
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
			return action(path)
		})
	} else {
		action(path)
	}

}
