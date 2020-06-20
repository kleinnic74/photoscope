package formats

import (
	"os"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

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
