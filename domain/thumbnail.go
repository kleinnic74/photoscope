package domain

import (
	"image"
	"io"
	"sort"
	"sync"

	"github.com/disintegration/gift"
)

var (
	Small  = ThumbSize{120, "S"}
	Medium = ThumbSize{427, "M"}
	Large  = ThumbSize{640, "L"}

	ThumbSizes = map[string]ThumbSize{
		Small.Name:  Small,
		Medium.Name: Medium,
		Large.Name:  Large,
	}
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
	return image.Image(orientation.Apply(thumb)), nil
}

type weightedThumber struct {
	thumber Thumber
	weight  float64
}

type weightedThumbers []weightedThumber

func (a weightedThumbers) Len() int           { return len(a) }
func (a weightedThumbers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a weightedThumbers) Less(i, j int) bool { return a[i].weight < a[j].weight }

type Thumbers struct {
	thumbers weightedThumbers

	lock sync.RWMutex
}

func (t *Thumbers) Add(thumber Thumber, weight float64) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.thumbers = append(t.thumbers, weightedThumber{thumber, weight})
	sort.Sort(t.thumbers)
}

func (t *Thumbers) CreateThumb(in io.Reader, format Format, orientation Orientation, size ThumbSize) (image.Image, error) {
	thumber := func() Thumber {
		t.lock.RLock()
		defer t.lock.RUnlock()
		return t.thumbers[0].thumber
	}()
	return thumber.CreateThumb(in, format, orientation, size)
}
