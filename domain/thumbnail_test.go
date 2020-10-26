package domain

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThumbnailSize(t *testing.T) {
	data := []struct {
		thumb ThumbSize
		in    image.Rectangle
		out   image.Rectangle
	}{
		{Medium, image.Rect(0, 0, 1920, 1080), image.Rect(0, 0, Medium.width, 1080*Medium.width/1920)},
		{Medium, image.Rect(0, 0, 1080, 1920), image.Rect(0, 0, 1080*Medium.width/1920, Medium.width)},
		{Small, image.Rect(0, 0, 500, 1000), image.Rect(0, 0, 500*Small.width/1000, Small.width)},
	}
	for i, d := range data {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			actual := d.thumb.BoundsOf(d.in)
			if actual != d.out {
				t.Errorf("Bad resulting thumb size %s: expected %s, got %s", d.thumb.Name, d.out, actual)
			}
		})
	}
}

func TestCreateThumb(t *testing.T) {
	in, err := os.Open("testdata/orientation/portrait_3.jpg")
	if err != nil {
		t.Fatalf("Failed to open test image: %s", err)
	}
	defer in.Close()
	var thumber LocalThumber
	result, err := thumber.CreateThumb(in, JPEG, 3, Medium)
	out, err := os.Create("testdata/orientation/portrait_3_thumb.jpg")
	if err != nil {
		t.Fatalf("FAiled to save thumb: %s", err)
	}
	defer out.Close()
	jpeg.Encode(out, result, &jpeg.Options{Quality: 75})
	assert.Equal(t, image.Rect(0, 0, 320, Medium.width), result.Bounds())
}
