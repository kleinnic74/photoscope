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

func imageResizer(format Format, in io.Reader, size ThumbSize) (image.Image, error) {
	image, err := format.Decode(in)
	if err != nil {
		return nil, err
	}
	return resize.Resize(size.width, 0, image, resize.NearestNeighbor), nil
}
