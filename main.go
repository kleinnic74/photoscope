// photos project main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
)

var (
	basedir        string
	matrixFilename string
	libDir         string
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s  <basedir>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&matrixFilename, "m", "distance.png", "Name of distance matrix file")
	flag.StringVar(&libDir, "l", "ile gophotos", "Path to photo library")
	flag.Parse()
	basedir := flag.Arg(0)
	if basedir == "" {
		flag.Usage()
		os.Exit(1)
	}
}

type Counter struct {
	count int
}

func (c *Counter) imageFound(img *domain.Photo) {
	log.Printf("Found photo: %s - Taken on: %s at %s", img.Filename, img.DateTaken, img.Location)
	c.count++
}

func (c *Counter) Total() int {
	return c.count
}

func main() {
	importer, err := NewDirectoryImporter(basedir)
	if err != nil {
		log.Fatalf("Cannot list photos from %s: %s", basedir, err)
	}
	//	classifier := NewEventClassifier()
	counter := Counter{}

	lib, err := library.NewBasicPhotoLibrary(libDir)
	if err != nil {
		log.Fatal(err)
	}
	err = importer.SkipDir("@eaDir").Walk(
		NewPhotoHandlerChain().Then(counter.imageFound).Then(lib.Add).Do())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d images", counter.Total())

	//	img := classifier.DistanceMatrixToImage()
	//	log.Printf("Creating time-distance matrix image %s", matrixFilename)
	//	out, err := os.Create(matrixFilename)
	//	if err != nil {
	//		log.Fatalf("Could not create distance matrix: %s", err)
	//	}
	//	defer out.Close()
	//	png.Encode(out, img)
}
