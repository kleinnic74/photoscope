package domain

import (
	"fmt"
	"io"

	"github.com/h2non/filetype"
)

const (
	Picture = iota
	Video
)

type Format struct {
	Type uint8
	Id   string
	Mime string
}

var (
	allFormats = []*Format{
		&Format{Type: Picture, Id: "jpg", Mime: "image/jpeg"},
		&Format{Type: Video, Id: "mov", Mime: "video/quicktime"},
	}

	formatsById map[string]*Format
)

func init() {
	formatsById = make(map[string]*Format)
	for _, f := range allFormats {
		formatsById[f.Id] = f
	}
}

func FormatForExt(ext string) (*Format, bool) {
	f, found := formatsById[ext]
	return f, found
}

func FormatOf(r io.Reader) (*Format, error) {
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
