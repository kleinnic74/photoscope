package domain

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
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
	data := []struct {
		Src    string
		Dst    string
		Bounds image.Rectangle
	}{
		{"testdata/orientation/portrait_0.jpg", "testdata/orientation/portrait_0_thumb.jpg", image.Rect(0, 0, 284, Medium.width)},
		{"testdata/orientation/portrait_1.jpg", "testdata/orientation/portrait_1_thumb.jpg", image.Rect(0, 0, 284, Medium.width)},
		{"testdata/orientation/portrait_2.jpg", "testdata/orientation/portrait_2_thumb.jpg", image.Rect(0, 0, 284, Medium.width)},
		{"testdata/orientation/portrait_3.jpg", "testdata/orientation/portrait_3_thumb.jpg", image.Rect(0, 0, 320, Medium.width)},
		{"testdata/orientation/portrait_4.jpg", "testdata/orientation/portrait_4_thumb.jpg", image.Rect(0, 0, 284, Medium.width)},
		{"testdata/orientation/portrait_5.jpg", "testdata/orientation/portrait_5_thumb.jpg", image.Rect(0, 0, 284, Medium.width)},
		{"testdata/orientation/portrait_6.jpg", "testdata/orientation/portrait_6_thumb.jpg", image.Rect(0, 0, 284, Medium.width)},
		{"testdata/orientation/portrait_7.jpg", "testdata/orientation/portrait_7_thumb.jpg", image.Rect(0, 0, 284, Medium.width)},
		{"testdata/orientation/portrait_8.jpg", "testdata/orientation/portrait_8_thumb.jpg", image.Rect(0, 0, 284, Medium.width)},
		// Landscape orientation (width > height)
		{"testdata/orientation/landscape_7.jpg", "testdata/orientation/landscape_7_thumb.jpg", image.Rect(0, 0, Medium.width, 284)},
	}
	for _, d := range data {
		t.Run(d.Src, func(t *testing.T) {
			in, err := os.Open(d.Src)
			if err != nil {
				t.Fatalf("Failed to open test image: %s", err)
			}
			var meta MediaMetaData
			if err := JPEG.DecodeMetaData(in, &meta); err != nil {
				t.Fatalf("Failed to decode metadata: %s", err)
			}
			in.Seek(0, 0)
			defer in.Close()
			var thumber LocalThumber
			result, err := thumber.CreateThumb(in, JPEG, meta.Orientation, Medium)
			dst := filepath.Join("testdata/out", filepath.Base(d.Dst))
			if err := saveImage(result, filepath.Join(dst)); err != nil {
				t.Fatalf("Failed to save %s: %s", dst, err)
			}
			assert.Equal(t, d.Bounds, result.Bounds())
		})
	}
}

func saveImage(img image.Image, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	jpeg.Encode(out, img, &jpeg.Options{Quality: 75})
	return nil
}
