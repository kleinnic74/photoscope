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
	data := []byte(`{
		"path": "2018/02/03",
		"id": "12345678",
		"format": "jpg",
		"gps": {
		   "long": 47.123445,
		   "lat": 45.12313
		}
	}`)
	var p Photo
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("Failed to Unmarshal JSON: %s", err)
	}
	assertEquals(t, "format", "jpg", p.Format().ID())
	assertEquals(t, "path", "2018/02/03", p.path)
	assertEquals(t, "id", "12345678", p.id)
	assertEquals(t, "gps.lat", "[45.123130;47.123445]", p.location.String())
}

func TestMarshallJSON(t *testing.T) {
	data := []struct {
		Photo Photo
		JSON  string
	}{
		{
			Photo: Photo{
				id:        "id",
				path:      "to/file",
				format:    domain.MustFormatForExt("jpg"),
				location:  gps.NewCoordinates(12, 34),
				dateTaken: time.Now(),
			},
			JSON: `{
  "path": "to/file",
  "id": "id",
  "format": "jpg",
  "dateUN": %d,
  "gps": {
    "lat": 12,
    "long": 34
  }
}`,
		},
	}
	for _, d := range data {
		dateUN := d.Photo.dateTaken.UnixNano()
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
			"jpg"),
			"2015/02/24",
			"myfile.jpg",
		},
	}
	for _, tt := range tests {
		actualPath, actualName, id := canonicalizeFilename(tt.photo)
		assertEquals(t, "name", tt.expectedName, actualName)
		assertEquals(t, "path", tt.expectedPath, actualPath)
		assertNotEmpty(t, "id", id)
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
	return gps.NewCoordinates((rand.Float64()-0.5)*360,
		(rand.Float64()-0.5)*90)
}
