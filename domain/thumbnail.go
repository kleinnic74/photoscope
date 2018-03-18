package domain

import (
	"image"

	"github.com/nfnt/resize"
)

var (
	Small  ThumbSize = ThumbSize{120, "S"}
	Medium ThumbSize = ThumbSize{427, "M"}
	Large  ThumbSize = ThumbSize{640, "L"}
)

type ThumbSize struct {
	long uint
	Name string
}

func Thumbnail(img image.Image, size ThumbSize) (image.Image, error) {
	return resize.Resize(size.long, 0, img, resize.NearestNeighbor), nil
}
