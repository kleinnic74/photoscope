package main

import (
	"image"
	"time"

	"image/draw"

	"image/color"

	"bitbucket.org/kleinnic74/photos/classification"
	"bitbucket.org/kleinnic74/photos/domain"
)

type EventClassifier struct {
	timestampedPhotos []classification.Timestamped
}

type TimestampedPhoto struct {
	domain.Photo
}

func (p *TimestampedPhoto) Timestamp() time.Time {
	return p.DateTaken()
}

func NewEventClassifier() *EventClassifier {
	return &EventClassifier{
		timestampedPhotos: make([]classification.Timestamped, 0),
	}
}

func (c *EventClassifier) Add(p domain.Photo) {
	tp := &TimestampedPhoto{p}
	c.timestampedPhotos = append(c.timestampedPhotos, tp)
}

func (c *EventClassifier) DistanceMatrixToImage() image.Image {
	kValues := []int{96, 48, 24, 6}
	size := len(c.timestampedPhotos)
	img := image.NewRGBA(image.Rect(0, 0, size*len(kValues), size))
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
	for n, k := range kValues {
		offset := n * size
		mat := classification.NewDistanceMatrixWithDistanceFunc(c.timestampedPhotos, classification.TimestampDistanceK(float64(k*3600)))
		for i := range mat {
			for j := range mat[i] {
				gray := uint8(mat[i][j] * 255)
				c := color.RGBA{gray, gray, gray, 255}
				img.Set(i+offset, j, c)
			}
		}
	}
	return img
}
