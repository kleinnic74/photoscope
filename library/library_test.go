package library

import (
	"fmt"
	"testing"

	"encoding/json"
	"math/rand"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJSON(t *testing.T) {
	data := []struct {
		json    []byte
		hasHash bool
	}{
		{[]byte(`{"path": "2018/02/03","id": "12345678","format": "jpg","gps": {"long": 47.123445,"lat": 45.12313}}`), false},
		{[]byte(`{"path": "2018/02/03","id": "12345678","format": "jpg","gps": {"long": 47.123445,"lat": 45.12313},"hash":"1234"}`), true},
	}
	for _, d := range data {
		var p Photo
		if err := json.Unmarshal(d.json, &p); err != nil {
			t.Fatalf("Failed to Unmarshal JSON: %s", err)
		}
		assertEquals(t, "format", "jpg", p.Format.ID())
		assertEquals(t, "path", "2018/02/03", p.Path)
		assertEquals(t, "id", "12345678", string(p.ID))
		assertEquals(t, "gps.lat", "[45.123130;47.123445]", p.Location.String())
		assert.Equal(t, d.hasHash, p.HasHash(), "Bad hash status")
	}
}

func TestMarshallJSON(t *testing.T) {
	data := []struct {
		Photo Photo
		JSON  string
	}{
		{
			Photo: Photo{
				ID:        "id",
				Path:      "to/file",
				Format:    domain.MustFormatForExt("jpg"),
				Location:  gps.MustNewCoordinates(12, 34),
				DateTaken: time.Now(),
				Hash:      BinaryHash("1234"),
			},
			JSON: `{
  "schema": 3,
  "id": "id",
  "path": "to/file",
  "format": "jpg",
  "size": 0,
  "dateUN": %d,
  "gps": {
    "lat": 12,
    "long": 34
  },
  "hash": "1234"
}`,
		},
	}
	for _, d := range data {
		dateUN := d.Photo.DateTaken.UnixNano()
		expected := fmt.Sprintf(d.JSON, dateUN)
		out, err := json.MarshalIndent(&d.Photo, "", "  ")
		if err != nil {
			t.Errorf("JSON marhsalling failed: %s", err)
		}
		assert.Equal(t, expected, string(out))
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	photo := RandomPhoto()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(photo)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestCanonicalizePhoto(t *testing.T) {
	var tests = []struct {
		photo        domain.Photo
		expectedPath string
		expectedName string
	}{
		{domain.NewPhotoFromFields("/some/path/myfile.jpg",
			at("2015", "02", "24"),
			somewhere(),
			"jpg", 1),
			"2015/02/24",
			"myfile.jpg",
		},
	}
	for _, tt := range tests {
		actualPath, actualName, id := canonicalizeFilename(tt.photo)
		assertEquals(t, "name", tt.expectedName, actualName)
		assertEquals(t, "path", tt.expectedPath, actualPath)
		assertNotEmpty(t, "id", string(id))
	}
}

func assertEquals(t *testing.T, name, expected, actual string) {
	if expected != actual {
		t.Errorf("Bad %s: expected '%s', got '%s'", name, expected, actual)
	}
}

func assertNotEmpty(t *testing.T, name, value string) {
	if len(value) == 0 {
		t.Errorf("Expected a non-empty string for '%s' but was empty", name)
	}
}

func at(year, month, day string) time.Time {
	t, err := time.Parse(time.RFC3339, fmt.Sprintf("%s-%s-%sT09:12:45Z", year, month, day))
	if err != nil {
		panic(err)
	}
	return t
}

func somewhere() *gps.Coordinates {
	coords, _ := gps.NewCoordinates((rand.Float64()-0.5)*360,
		(rand.Float64()-0.5)*90)
	return coords
}
