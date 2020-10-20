package domain

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/h2non/filetype"
	"github.com/rwcarlsen/goexif/exif"

	"bitbucket.org/kleinnic74/photos/domain/formats"
	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type MediaType uint8

const (
	// Picture is the Format type value for pictures (or images)
	Picture = MediaType(iota)
	// Video is the Format type value for videos
	Video = MediaType(iota)
)

type metaDataReader func(io.Reader, *MediaMetaData) error
type photoDecoder func(io.Reader) (image.Image, error)
type photoEncoder func(image.Image, io.Writer) error
type thumbFunc func(io.Reader) (image.Image, error)

type Format interface {
	Type() MediaType
	ID() string
	Mime() string
	DecodeMetaData(in io.Reader, meta *MediaMetaData) error
	Decode(in io.Reader) (image.Image, error)
	Encode(img image.Image, out io.Writer) error
	Thumbbase(in io.Reader) (image.Image, error)
}

type formatImpl struct {
	typeID     MediaType
	id         string
	mime       string
	metaReader metaDataReader
	decoder    photoDecoder
	encoder    photoEncoder
	thumber    thumbFunc
}

type ErrThumbsNotSupported string

func (e ErrThumbsNotSupported) Error() string {
	return fmt.Sprintf("No thumbs available for %s", string(e))
}

var (
	formatsById map[string]Format = map[string]Format{}

	ErrNoDecoderAvailable = errors.New("No decoder available for this format")
	ErrNoEncoderAvailable = errors.New("No encoder available for this format")
)

var noopFormat = formatImpl{
	typeID:     Picture,
	metaReader: func(io.Reader, *MediaMetaData) error { return nil },
}

var (
	JPEG Format
	MOV  Format
)

func init() {
	JPEG = RegisterFormat(Picture, "jpg", "image/jpeg", exifReader, jpeg.Decode, jpegEncode, jpeg.Decode)
	MOV = RegisterFormat(Video, "mov", "video/quicktime", quicktimeReader, nil, nil, nil)
}

func RegisterFormat(typeID MediaType, extension string, mime string,
	metaReader metaDataReader,
	decoder photoDecoder,
	encoder photoEncoder,
	thumber thumbFunc) (format Format) {
	format = formatImpl{
		typeID:     typeID,
		id:         extension,
		mime:       mime,
		metaReader: metaReader,
		decoder:    decoder,
		encoder:    encoder,
		thumber:    thumber,
	}
	formatsById[extension] = format
	return
}

func FormatForExt(ext string) (Format, bool) {
	f, found := formatsById[ext]
	return f, found
}

func MustFormatForExt(ext string) Format {
	if ext == "" {
		return noopFormat
	}
	f, found := formatsById[ext]
	if !found {
		panic(fmt.Errorf("Unkown format with extension '%s'", ext))
	}
	return f
}

// FormatOf returns the format of the image in the given reader. Calling
// this function will consume the reader
func FormatOf(r io.Reader) (Format, error) {
	header := make([]byte, 500)
	r.Read(header)
	kind, err := filetype.Match(header)
	if err != nil {
		return nil, err
	}
	if f, found := formatsById[kind.Extension]; found {
		return f, nil
	} else {
		return nil, fmt.Errorf("Unsupported file format")
	}
}

func (f formatImpl) Type() MediaType {
	return f.typeID
}

func (f formatImpl) ID() string {
	return f.id
}

func (f formatImpl) Mime() string {
	return f.mime
}

func (f formatImpl) Thumbbase(in io.Reader) (image.Image, error) {
	if f.thumber == nil {
		return nil, ErrThumbsNotSupported(f.id)
	}
	return f.thumber(in)
}

// DecodeMetaData will decode meta-data as per this format from the given
// reader and store it in the given metadata instance
func (f formatImpl) DecodeMetaData(in io.Reader, meta *MediaMetaData) error {
	if f.metaReader != nil {
		return f.metaReader(in, meta)
	}
	return nil
}

// Decode decodes the binary data from the given reader as an image in this format
func (f formatImpl) Decode(in io.Reader) (image.Image, error) {
	if f.decoder == nil {
		return nil, ErrNoDecoderAvailable
	}
	return f.decoder(in)
}

// Encode encodes this image in the current format into the given writer
func (f formatImpl) Encode(img image.Image, out io.Writer) error {
	if f.encoder == nil {
		return ErrNoEncoderAvailable
	}
	return f.encoder(img, out)
}

func exifReader(in io.Reader, meta *MediaMetaData) error {
	ex, err := exif.Decode(in)
	if err != nil {
		return err
	}
	if dateTaken, err := ex.DateTime(); err == nil {
		meta.DateTaken = dateTaken
	}
	if lat, long, err := ex.LatLong(); err == nil {
		meta.Location = gps.NewCoordinates(lat, long)
	}
	return nil
}

func quicktimeReader(in io.Reader, meta *MediaMetaData) error {
	qt, err := formats.ReadAsQuicktime(in)
	if err != nil {
		return err
	}
	meta.DateTaken = qt.DateTaken()
	meta.Location = qt.Location()
	return nil
}

func jpegEncode(img image.Image, out io.Writer) error {
	return jpeg.Encode(out, img, nil)
}
