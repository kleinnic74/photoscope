package domain

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"

	"github.com/h2non/filetype"
	"github.com/rwcarlsen/goexif/exif"

	"bitbucket.org/kleinnic74/photos/domain/formats"
	"bitbucket.org/kleinnic74/photos/domain/gps"
)

const (
	// Picture is the Format type value for pictures (or images)
	Picture = iota
	// Video is the Format type value for videos
	Video
)

type metaDataReader func(io.Reader, *MediaMetaData) error
type photoDecoder func(io.Reader) (image.Image, error)
type photoEncoder func(image.Image, io.Writer)

type Format struct {
	Type       uint8
	Id         string
	Mime       string
	metaReader metaDataReader
	decoder    photoDecoder
	encoder    photoEncoder
}

var (
	allFormats = []*Format{
		&Format{Type: Picture, Id: "jpg", Mime: "image/jpeg", metaReader: exifReader, decoder: jpeg.Decode, encoder: jpegEncode},
		&Format{Type: Video, Id: "mov", Mime: "video/quicktime", metaReader: quicktimeReader},
	}

	formatsById map[string]*Format
)

func init() {
	formatsById = make(map[string]*Format)
	for _, f := range allFormats {
		formatsById[f.Id] = f
	}
}

func FormatForExt(ext string) (*Format, bool) {
	f, found := formatsById[ext]
	return f, found
}

func MustFormatForExt(ext string) *Format {
	f, found := formatsById[ext]
	if !found {
		panic(fmt.Errorf("Unkown format with extension '%s'", ext))
	}
	return f
}

// FormatOf returns the format of the image in the given reader. Calling
// this function will consume the reader
func FormatOf(r io.Reader) (*Format, error) {
	header := make([]byte, 500)
	r.Read(header)
	kind, err := filetype.Match(header)
	if err != nil {
		return nil, err
	}
	if f, found := formatsById[kind.Extension]; found {
		return f, nil
	} else {
		return nil, fmt.Errorf("Unsupported file format")
	}
}

// DecodeMetaData will decode meta-data as per this format from the given
// reader and store it in the given metadata instance
func (f *Format) DecodeMetaData(in io.Reader, meta *MediaMetaData) error {
	if f.metaReader != nil {
		return f.metaReader(in, meta)
	}
	return nil
}

// Decode decodes the binary data from the given reader as an image in this format
func (f *Format) Decode(in io.Reader) (image.Image, error) {
	return f.decoder(in)
}

// Encode encodes this image in the current format into the given writer
func (f *Format) Encode(img image.Image, out io.Writer) {
	f.encoder(img, out)
}

func exifReader(in io.Reader, meta *MediaMetaData) error {
	ex, err := exif.Decode(in)
	if err != nil {
		return err
	}
	if dateTaken, err := ex.DateTime(); err == nil {
		meta.DateTaken = dateTaken
	}
	if lat, long, err := ex.LatLong(); err == nil {
		c := gps.NewCoordinates(lat, long)
		meta.Location = &c
	}
	return nil
}

func quicktimeReader(in io.Reader, meta *MediaMetaData) error {
	qt, err := formats.ReadAsQuicktime(in)
	if err != nil {
		return err
	}
	meta.DateTaken = qt.DateTaken()
	meta.Location = qt.Location()
	return nil
}

func jpegEncode(img image.Image, out io.Writer) {
	if err := jpeg.Encode(out, img, nil); err != nil {
		log.Printf("Error while encoding to jpeg: %s", err)
	}
}
