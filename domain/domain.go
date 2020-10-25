package domain

import (
	"image"
	"image/draw"
	"io"
	"os"
	"time"

	"path/filepath"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/disintegration/gift"
)

// Identifiable represents any object that can be uniquely identified with an string ID
type Identifiable interface {
	ID() string
}

// Orientation represents the EXIF orientation of the image
type Orientation int

const (
	UnknownOrientation = Orientation(iota)
	NormalOrientation
	FlipHorizontalOrientation
	Rotate180Orientation
	FlipVerticalOrientation
	TransposeOrientation
	Rotate270Orientation
	TransverseOrientation
	Rotate90Orientation
)

// Filters to rotate an image according to its EXIF orientation tag
//  see https://storage.googleapis.com/go-attachment/4341/0/exif-orientations.png
var orientationFilters = []gift.Filter{
	gift.FlipHorizontal(), // EXIF 2
	gift.Rotate180(),
	gift.FlipVertical(),
	gift.Transpose(),
	gift.Rotate270(),
	gift.Transverse(),
	gift.Rotate90(), // EXIF 8
}

func (o Orientation) Apply(src image.Image) image.Image {
	filter, needsTransform := o.Filter()
	if !needsTransform {
		return src
	}
	dst := image.NewRGBA(src.Bounds())
	filter.Draw(dst, src, nil)
	return image.Image(dst)
}

func (o Orientation) Filter() (gift.Filter, bool) {
	i := int(o) - 2
	switch {
	case i < 0 || i >= len(orientationFilters):
		return nil, false
	default:
		return orientationFilters[i], true
	}
}

type noopFilter struct{}

func (f noopFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
}

func (f noopFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	return srcBounds
}

// MediaMetaData contains meta-information about a media object
type MediaMetaData struct {
	DateTaken   time.Time
	Location    *gps.Coordinates
	Orientation Orientation
}

// Photo represents one image in a media library
type Photo interface {
	Identifiable
	Name() string
	Format() FormatSpec
	SizeInBytes() int64
	Content() (img io.ReadCloser, err error)
	Image() (image.Image, error)

	DateTaken() time.Time
	Location() *gps.Coordinates
	Orientation() Orientation
}

type photoFile struct {
	filename    string
	path        string
	size        int64
	dateTaken   time.Time
	format      FormatSpec
	location    *gps.Coordinates
	orientation Orientation
}

// NewPhoto creates a new Photo instance from the image file at the given path
func NewPhoto(path string) (Photo, error) {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	format, err := FormatOf(f)
	if err != nil {
		return nil, err
	}
	f.Seek(0, io.SeekStart)
	meta := guessMeta(fileinfo)
	if err = format.DecodeMetaData(f, meta); err != nil {
		return nil, err
	}
	return &photoFile{
		filename:    filenameFromPath(path),
		path:        path,
		size:        fileinfo.Size(),
		dateTaken:   meta.DateTaken,
		location:    meta.Location,
		orientation: meta.Orientation,
		format:      format,
	}, nil
}

func guessMeta(fileinfo os.FileInfo) *MediaMetaData {
	return &MediaMetaData{
		DateTaken: fileinfo.ModTime(),
		Location:  gps.Unknown,
	}
}

func NewPhotoFromFields(path string, taken time.Time, location *gps.Coordinates, format string, orientation Orientation) Photo {
	fullpath := filenameFromPath(path)
	return &photoFile{
		filename:    fullpath,
		path:        path,
		size:        0,
		dateTaken:   taken,
		location:    location,
		format:      MustFormatForExt(format),
		orientation: orientation,
	}
}

func filenameFromPath(path string) string {
	ext := filepath.Ext(path)
	filename := filepath.Base(path)
	filename = filename[:len(filename)-len(ext)]
	return filename
}

func (p *photoFile) ID() string {
	return p.filename
}

func (p *photoFile) Name() string {
	return p.filename
}

func (p *photoFile) DateTaken() time.Time {
	return p.dateTaken
}

func (p *photoFile) Format() FormatSpec {
	return p.format
}

func (p *photoFile) Location() *gps.Coordinates {
	return p.location
}

func (p *photoFile) Orientation() Orientation {
	return p.orientation
}

func (p *photoFile) Image() (image.Image, error) {
	in, err := p.Content()
	if err != nil {
		return nil, err
	}
	defer in.Close()
	return p.format.Decode(in)
}

func (p *photoFile) Content() (io.ReadCloser, error) {
	return os.Open(p.path)
}

func (p *photoFile) SizeInBytes() int64 {
	return p.size
}
