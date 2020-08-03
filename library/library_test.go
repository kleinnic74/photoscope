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
	p := Photo{
		id:        "id",
		path:      "to/file",
		format:    domain.MustFormatForExt("jpg"),
		location:  gps.NewCoordinates(12, 34),
		dateTaken: time.Now(),
	}
	dateUN := p.dateTaken.UnixNano()
	assert.Equal(t, "[12.000000;34.000000]", p.location.String())
	expected := `{
  "path": "to/file",
  "id": "id",
  "format": "jpg",
  "dateUN": %d,
  "gps": {
    "lat": 12,
    "long": 34
  }
}`
	expected = fmt.Sprintf(expected, dateUN)
	out, err := json.MarshalIndent(&p, "", "  ")
	if err != nil {
		t.Errorf("JSON marhsalling failed: %s", err)
	}
	if expected != string(out) {
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

func somewhere() gps.Coordinates {
	return gps.NewCoordinates((rand.Float64()-0.5)*360,
		(rand.Float64()-0.5)*90)
}
