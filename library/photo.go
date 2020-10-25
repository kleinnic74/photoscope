// photo.go
package library

import (
	"encoding/json"
	"path/filepath"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/gps"
)

const currentSchema = 1

type Photo struct {
	schema      uint
	ID          PhotoID            `json:"id"`
	Path        string             `json:"path"`
	Size        int64              `json:"size"`
	Orientation domain.Orientation `json:"or,omitempty"`
	Format      domain.FormatSpec  `json:"format"`
	DateTaken   time.Time          `json:"dateUN,omitempty"`
	Location    *gps.Coordinates   `json:"gps,omitempty"`
	Hash        BinaryHash         `json:"hash,omitempty"`
}

func (p *Photo) Name() string {
	return filepath.Base(p.Path)
}

func (p *Photo) HasHash() bool {
	return len(p.Hash) > 0
}

func (p *Photo) MarshalJSON() ([]byte, error) {
	out := struct {
		Schema      uint               `json:"schema"`
		ID          PhotoID            `json:"id"`
		Path        string             `json:"path"`
		Format      string             `json:"format"`
		Size        int                `json:"size"`
		DateTaken   int64              `json:"dateUN"`
		Location    *gps.Coordinates   `json:"gps"`
		Orientation domain.Orientation `json:"or,omitempty"`
	}{
		Schema:      currentSchema,
		ID:          p.ID,
		Path:        p.Path,
		Format:      p.Format.ID(),
		DateTaken:   p.DateTaken.UnixNano(),
		Location:    p.Location,
		Orientation: p.Orientation,
	}
	return json.Marshal(&out)
}

func (p *Photo) UnmarshalJSON(buf []byte) error {
	// TODO get rid of this, format should be marshallabled to string
	var data struct {
		Schema      uint               `json:"schema"`
		Path        string             `json:"path"`
		ID          PhotoID            `json:"id"`
		Format      string             `json:"format"`
		Size        int                `json:"size"`
		DateTaken   int64              `json:"dateUN"`
		Location    *gps.Coordinates   `json:"gps"`
		Orientation domain.Orientation `json:"or,omitempty"`
	}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		return err
	}
	p.ID = data.ID
	p.Format = domain.MustFormatForExt(data.Format)
	p.Path = data.Path
	if data.DateTaken != 0 {
		p.DateTaken = time.Unix(data.DateTaken/1e9, data.DateTaken%1e9)
	}
	p.Location = data.Location
	p.Orientation = data.Orientation
	p.schema = data.Schema
	return nil
}
