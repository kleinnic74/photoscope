package domain

import (
	"io"
	"log"
	"os"
	"time"

	"path/filepath"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

type MediaMetaData struct {
	DateTaken time.Time
	Location  *gps.Coordinates
}

type Photo interface {
	Id() string
	Format() *Format
	Content() (io.ReadCloser, error)
	DateTaken() time.Time
	Location() *gps.Coordinates
}

type photoFile struct {
	filename  string
	path      string
	dateTaken time.Time
	format    *Format
	location  *gps.Coordinates
}

type TagHandler func(name, value string)

type exifWalker struct {
	w TagHandler
}

func (w *exifWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	w.w(string(name), tag.String())
	return nil
}

func PrintExif(path string, walker func(name, value string)) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	meta, err := exif.Decode(f)
	if err != nil {
		return err
	}
	return meta.Walk(&exifWalker{w: walker})
}

func NewPhoto(path string) (Photo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	format, err := FormatOf(f)
	if err != nil {
		return nil, err
	}
	f.Seek(0, io.SeekStart)
	meta := guessMeta(path)
	if err = format.DecodeMetaData(f, meta); err != nil {
		log.Printf("  Error while decoding meta-data: %s", err)
	}
	return &photoFile{
		filename:  filenameFromPath(path),
		path:      path,
		dateTaken: meta.DateTaken,
		location:  meta.Location,
		format:    format,
	}, nil
}

func guessMeta(path string) *MediaMetaData {
	fileinfo, _ := os.Stat(path)
	return &MediaMetaData{
		DateTaken: fileinfo.ModTime(),
		Location:  gps.Unknown,
	}
}

func NewPhotoFromFields(path string, taken time.Time, location gps.Coordinates, format string) Photo {
	return &photoFile{
		filename:  filenameFromPath(path),
		path:      path,
		dateTaken: taken,
		location:  &location,
		format:    MustFormatForExt(format),
	}
}

func filenameFromPath(path string) string {
	ext := filepath.Ext(path)
	filename := filepath.Base(path)
	filename = filename[:len(filename)-len(ext)]
	return filename
}

func (p *photoFile) Id() string {
	return p.filename
}

func (p *photoFile) DateTaken() time.Time {
	return p.dateTaken
}

func (p *photoFile) Format() *Format {
	return p.format
}

func (p *photoFile) Location() *gps.Coordinates {
	return p.location
}

func (p *photoFile) Content() (io.ReadCloser, error) {
	return os.Open(p.path)
}
