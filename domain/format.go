package domain

import (
	"fmt"
	"io"

	"github.com/h2non/filetype"
	"github.com/rwcarlsen/goexif/exif"

	"bitbucket.org/kleinnic74/photos/domain/formats"
	"bitbucket.org/kleinnic74/photos/domain/gps"
)

const (
	Picture = iota
	Video
)

type metaDataReader func(io.Reader, *MediaMetaData) error

type Format struct {
	Type       uint8
	Id         string
	Mime       string
	metaReader metaDataReader
}

var (
	allFormats = []*Format{
		&Format{Type: Picture, Id: "jpg", Mime: "image/jpeg", metaReader: exifReader},
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

func (f *Format) DecodeMetaData(in io.Reader, meta *MediaMetaData) error {
	if f.metaReader != nil {
		return f.metaReader(in, meta)
	}
	return nil
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
