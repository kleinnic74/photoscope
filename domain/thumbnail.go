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
	cost    float64
}

type byAscendingCosts []weightedThumber

func (a byAscendingCosts) Len() int           { return len(a) }
func (a byAscendingCosts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byAscendingCosts) Less(i, j int) bool { return a[i].cost < a[j].cost }

type Thumbers struct {
	thumbers []weightedThumber

	lock sync.RWMutex
}

func (t *Thumbers) Add(thumber Thumber, cost float64) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.thumbers = append(t.thumbers, weightedThumber{thumber, cost})
	sort.Sort(byAscendingCosts(t.thumbers))
}

func (t *Thumbers) CreateThumb(in io.Reader, format Format, orientation Orientation, size ThumbSize) (img image.Image, err error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, thumber := range t.thumbers {
		img, err = thumber.thumber.CreateThumb(in, format, orientation, size)
		if err == nil {
			return
		}
	}
	return
}
