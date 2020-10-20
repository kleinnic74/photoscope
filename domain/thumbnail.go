package domain

import (
	"image"
	"io"

	"github.com/nfnt/resize"
)

var (
	Small  ThumbSize = ThumbSize{120, "S"}
	Medium ThumbSize = ThumbSize{427, "M"}
	Large  ThumbSize = ThumbSize{640, "L"}
)

type ThumbSize struct {
	width uint
	Name  string
}

type Thumber interface {
	CreateThumb(Format, io.Reader, ThumbSize) (image.Image, error)
}

type LocalThumber struct{}

func (t LocalThumber) CreateThumb(format Format, in io.Reader, size ThumbSize) (image.Image, error) {
	img, err := format.Thumbbase(in)
	if err != nil {
		return nil, err
	}
	return resize.Thumbnail(size.width, size.width, img, resize.Bilinear), nil
}
