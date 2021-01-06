// photo.go
package library

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/reusee/mmh3"
)

const currentSchema = 6

type Photo struct {
	ExtendedPhotoID
	schema      Version
	Path        string             `json:"path,omitempty"`
	Size        int64              `json:"size,omitempty"`
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
		ExtendedPhotoID
		Schema      Version            `json:"schema"`
		Path        string             `json:"path,omitempty"`
		Format      string             `json:"format"`
		Size        int                `json:"size,omitempty"`
		DateTaken   int64              `json:"dateUN"`
		Location    *gps.Coordinates   `json:"gps,omitempty"`
		Orientation domain.Orientation `json:"or,omitempty"`
		Hash        BinaryHash         `json:"hash,omitempty"`
	}{
		Schema:          currentSchema,
		ExtendedPhotoID: p.ExtendedPhotoID,
		Path:            p.Path,
		Format:          p.Format.ID(),
		DateTaken:       p.DateTaken.UnixNano(),
		Location:        p.Location,
		Orientation:     p.Orientation,
		Hash:            p.Hash,
	}
	return json.Marshal(&out)
}

func (p *Photo) UnmarshalJSON(buf []byte) error {
	// TODO get rid of this, format should be marshallabled to string
	var data struct {
		ExtendedPhotoID
		Schema      Version            `json:"schema"`
		Path        string             `json:"path"`
		Format      domain.FormatSpec  `json:"format"`
		Size        int                `json:"size"`
		DateTaken   int64              `json:"dateUN"`
		Location    *gps.Coordinates   `json:"gps"`
		Orientation domain.Orientation `json:"or,omitempty"`
		Hash        BinaryHash         `json:"hash,omitempty"`
	}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		return err
	}
	p.ExtendedPhotoID = data.ExtendedPhotoID
	p.Format = data.Format
	p.Path = data.Path
	if data.DateTaken != 0 {
		p.DateTaken = time.Unix(data.DateTaken/1e9, data.DateTaken%1e9).In(time.UTC)
	}
	p.Location = data.Location
	p.Orientation = data.Orientation
	p.Hash = data.Hash
	p.schema = data.Schema
	return nil
}

func orderedIDOf(ts time.Time, key PhotoID) OrderedID {
	var id bytes.Buffer
	id.Write([]byte(ts.UTC().Format(time.RFC3339)))
	h := mmh3.New32()
	h.Write([]byte(strings.ToLower(string(key))))
	id.Write(h.Sum(nil))
	return OrderedID(id.Bytes())
}

func boundaryIDs(begin, end time.Time) (low, high OrderedID) {
	var lbuf bytes.Buffer
	lbuf.Write([]byte(begin.UTC().Format(time.RFC3339)))
	lbuf.Write([]byte{0, 0, 0, 0})
	low = OrderedID(lbuf.Bytes())
	var hbuf bytes.Buffer
	hbuf.Write([]byte(end.UTC().Format(time.RFC3339)))
	hbuf.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	high = OrderedID(lbuf.Bytes())
	return
}
