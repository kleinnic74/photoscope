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

// MediaMetaData contains meta-information about a media object
type MediaMetaData struct {
	DateTaken time.Time
	Location  *gps.Coordinates
}

// Photo represents one image in a media library
type Photo interface {
	Id() string
	Format() *Format
	SizeInBytes() int64
	Content() (img io.ReadCloser, err error)
	Image() (image.Image, error)
	Thumb(ThumbSize) (img image.Image, err error)

	DateTaken() time.Time
	Location() *gps.Coordinates
}

type photoFile struct {
	filename  string
	path      string
	size      int64
	dateTaken time.Time
	format    *Format
	location  *gps.Coordinates
}

// NewPhoto creates a new Photo instance from the image file at the given path
func NewPhoto(path string) (Photo, error) {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
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
	meta := guessMeta(fileinfo)
	if err = format.DecodeMetaData(f, meta); err != nil {
		log.Printf("  Error while decoding meta-data: %s", err)
	}
	return &photoFile{
		filename:  filenameFromPath(path),
		path:      path,
		size:      fileinfo.Size(),
		dateTaken: meta.DateTaken,
		location:  meta.Location,
		format:    format,
	}, nil
}

func guessMeta(fileinfo os.FileInfo) *MediaMetaData {
	return &MediaMetaData{
		DateTaken: fileinfo.ModTime(),
		Location:  gps.Unknown,
	}
}

func NewPhotoFromFields(path string, taken time.Time, location gps.Coordinates, format string) (Photo, error) {
	fullpath := filenameFromPath(path)
	var size int64 = -1
	if fileinfo, err := os.Stat(fullpath); err == nil {
		size = fileinfo.Size()
	}
	return &photoFile{
		filename:  fullpath,
		path:      path,
		size:      size,
		dateTaken: taken,
		location:  &location,
		format:    MustFormatForExt(format),
	}, nil
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

func (p *photoFile) SizeInBytes() int64 {
	return p.size
}
