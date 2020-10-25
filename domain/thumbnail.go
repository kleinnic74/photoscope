package domain

import (
	"image"
	"io"

	"github.com/disintegration/gift"
)

var (
	Small  = ThumbSize{120, "S"}
	Medium = ThumbSize{427, "M"}
	Large  = ThumbSize{640, "L"}
)

type ThumbSize struct {
	width int
	Name  string
}

func (size ThumbSize) BoundsOf(img image.Rectangle) image.Rectangle {
	if img.Dx() > img.Dy() {
		return image.Rect(0, 0, size.width, (size.width*img.Dy())/img.Dx())
	} else {
		return image.Rect(0, 0, (size.width*img.Dx())/img.Dy(), size.width)
	}
}

type Thumber interface {
	CreateThumb(io.Reader, Format, Orientation, ThumbSize) (image.Image, error)
}

type LocalThumber struct{}

func (t LocalThumber) CreateThumb(in io.Reader, format Format, orientation Orientation, size ThumbSize) (image.Image, error) {
	img, err := format.Thumbbase(in)
	if err != nil {
		return nil, err
	}
	targetSize := size.BoundsOf(img.Bounds())
	thumb := image.NewRGBA(targetSize)
	filter := gift.New(
		gift.ResizeToFit(targetSize.Dx(), targetSize.Dy(), gift.LinearResampling),
	)
	filter.Draw(thumb, img)
	if reorientate, needed := orientation.Filter(); needed {
		targetSize = filter.Bounds(targetSize)
		rotated := image.NewRGBA(targetSize)
		reorientate.Draw(rotated, thumb, nil)
		thumb = rotated
	}
	return thumb, nil
}
