// photo.go
package library

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"math/rand"
	"path/filepath"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type Photo struct {
	lib       *BasicPhotoLibrary
	path      string
	size      int64
	id        string
	format    domain.Format
	dateTaken time.Time
	location  gps.Coordinates
}

func (p *Photo) ID() string {
	return p.id
}

func (p *Photo) Name() string {
	return filepath.Base(p.path)
}

func (p *Photo) Format() domain.Format {
	return p.format
}

func (p *Photo) DateTaken() time.Time {
	return p.dateTaken
}

func (p *Photo) Location() *gps.Coordinates {
	return &p.location
}

func (p *Photo) Content() (io.ReadCloser, error) {
	return p.lib.openPhoto(p.path)
}

func (p *Photo) SizeInBytes() int64 {
	if p.size == -1 {
		p.size = p.lib.fileSizeOf(p.path)
	}
	return p.size
}

func (p *Photo) Image() (image.Image, error) {
	content, err := p.Content()
	if err != nil {
		return nil, err
	}
	return p.format.Decode(content)
}

func (p *Photo) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Path      string          `json:"path"`
		ID        string          `json:"id"`
		Format    string          `json:"format"`
		DateTaken int64           `json:"dateUN,omitempty"`
		Location  gps.Coordinates `json:"gps,omitempty"`
	}{
		Path:      p.path,
		ID:        p.id,
		Format:    p.format.ID(),
		DateTaken: p.dateTaken.UnixNano(),
		Location:  p.location,
	})
}

func (p *Photo) UnmarshalJSON(buf []byte) error {
	data := struct {
		Path      string          `json:"path"`
		ID        string          `json:"id"`
		Format    string          `json:"format"`
		DateTaken int64           `json:"dateUN"`
		Location  gps.Coordinates `json:"gps"`
	}{}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		return err
	}
	p.id = data.ID
	p.format = domain.MustFormatForExt(data.Format)
	p.path = data.Path
	p.size = -1
	p.dateTaken = time.Unix(data.DateTaken/1e9, data.DateTaken%1e9)
	p.location = data.Location
	return nil
}

func RandomPhoto() *Photo {
	id := rand.Uint32()
	dateTaken, _ := time.Parse(time.RFC3339, "2018-02-23T13:43:12Z")
	f := domain.MustFormatForExt("jpg")
	return &Photo{
		id:        fmt.Sprintf("%8d", id),
		path:      "2018/02/23",
		format:    f,
		dateTaken: dateTaken,
		location:  gps.NewCoordinates(47.123445, 45.12313),
	}
}
