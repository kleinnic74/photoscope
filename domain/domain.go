package domain

import (
	"fmt"
	"os"
	"time"

	"path/filepath"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

type Coordinates struct {
	lat  float64
	long float64
}

func (c Coordinates) String() string {
	return fmt.Sprintf("[%f;%f]", c.lat, c.long)
}

type Photo struct {
	Filename  string
	Path      string
	DateTaken time.Time
	Location  *Coordinates
	Format    string
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

func NewPhoto(path string) (*Photo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var (
		taken time.Time
		gps   *Coordinates
	)
	filename := filepath.Base(path)
	format := filepath.Ext(path)

	meta, err := exif.Decode(f)
	if err == nil {
		taken, err = meta.DateTime()
		if lat, long, err := meta.LatLong(); err == nil {
			gps = &Coordinates{lat, long}
		}
	}
	if err != nil {
		fileinfo, _ := os.Stat(path)
		taken = fileinfo.ModTime()
	}
	return &Photo{
		Filename:  filename,
		Path:      path,
		DateTaken: taken,
		Location:  gps,
		Format:    format,
	}, nil
}

func (p *Photo) Timestamp() time.Time {
	return p.DateTaken
}
