package classification

import (
	"image"
	"time"

	"image/draw"

	"image/color"

	"bitbucket.org/kleinnic74/photos/library"
)

type tsPhoto struct {
	photo *library.Photo
}

func (p tsPhoto) Timestamp() time.Time {
	return p.photo.DateTaken
}

type EventClassifier struct {
	timestampedPhotos []Timestamped
}

func NewEventClassifier() *EventClassifier {
	return &EventClassifier{}
}

func (c *EventClassifier) Add(p *library.Photo) {
	c.timestampedPhotos = append(c.timestampedPhotos, tsPhoto{p})
}

func (c *EventClassifier) DistanceMatrixToImage() image.Image {
	kValues := []time.Duration{96, 48, 24, 6}
	size := len(c.timestampedPhotos)
	img := image.NewRGBA(image.Rect(0, 0, size*len(kValues), size))
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
	for n, k := range kValues {
		offset := n * size
		mat := NewDistanceMatrixWithDistanceFunc(TimestampDistance(k * time.Hour))
		ssm := mat.SelfSimilarityMatrix(c.timestampedPhotos)
		for i := range ssm {
			for j := range ssm[i] {
				gray := uint8(ssm[i][j] * 255)
				c := color.RGBA{gray, gray, gray, 255}
				img.Set(i+offset, j, c)
			}
		}
	}
	return img
}
