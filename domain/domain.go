package domain

import (
	"image"
	"io"
	"log"
	"os"
	"time"

	"path/filepath"

	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type MediaMetaData struct {
	DateTaken time.Time
	Location  *gps.Coordinates
}

type Photo interface {
	Id() string
	Format() *Format
	Content() (io.ReadCloser, error)
	Image() (image.Image, error)
	Thumb(ThumbSize) (image.Image, error)

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

func (p *photoFile) Image() (image.Image, error) {
	in, err := p.Content()
	if err != nil {
		return nil, err
	}
	defer in.Close()
	return p.format.Decode(in)
}

func (p *photoFile) Thumb(size ThumbSize) (image.Image, error) {
	content, err := p.Content()
	if err != nil {
		return nil, err
	}
	defer content.Close()
	img, err := p.format.Decode(content)
	if err != nil {
		return nil, err
	}
	return Thumbnail(img, size)
}

func (p *photoFile) Content() (io.ReadCloser, error) {
	return os.Open(p.path)
}
