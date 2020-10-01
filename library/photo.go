// photo.go
package library

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type Photo struct {
	ID        string           `json:"id"`
	Path      string           `json:"path"`
	Size      int64            `json:"size"`
	Format    domain.Format    `json:"format"`
	DateTaken time.Time        `json:"dateUN,omitempty" storm:"index"`
	Location  *gps.Coordinates `json:"gps,omitempty"`
}

func (p *Photo) Name() string {
	return filepath.Base(p.Path)
}

func (p *Photo) MarshalJSON() ([]byte, error) {
	out := struct {
		ID        string           `json:"id"`
		Path      string           `json:"path"`
		Format    string           `json:"format"`
		DateTaken int64            `json:"dateUN"`
		Location  *gps.Coordinates `json:"gps"`
	}{
		ID:        p.ID,
		Path:      p.Path,
		Format:    p.Format.ID(),
		DateTaken: p.DateTaken.UnixNano(),
		Location:  p.Location,
	}
	return json.Marshal(&out)
}

func (p *Photo) UnmarshalJSON(buf []byte) error {
	// TODO get rid of this, format should be marshallabled to string
	data := struct {
		Path      string           `json:"path"`
		ID        string           `json:"id"`
		Format    string           `json:"format"`
		DateTaken int64            `json:"dateUN"`
		Location  *gps.Coordinates `json:"gps"`
	}{}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		return err
	}
	p.ID = data.ID
	p.Format = domain.MustFormatForExt(data.Format)
	p.Path = data.Path
	p.Size = -1
	p.DateTaken = time.Unix(data.DateTaken/1e9, data.DateTaken%1e9)
	p.Location = data.Location
	return nil
}

func RandomPhoto() *Photo {
	id := rand.Uint32()
	dateTaken, _ := time.Parse(time.RFC3339, "2018-02-23T13:43:12Z")
	f := domain.MustFormatForExt("jpg")
	return &Photo{
		ID:        fmt.Sprintf("%8d", id),
		Path:      "2018/02/23",
		Format:    f,
		DateTaken: dateTaken,
		Location:  gps.NewCoordinates(47.123445, 45.12313),
	}
}
