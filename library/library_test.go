package library

import (
	"fmt"
	"path/filepath"
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
				ExtendedPhotoID: ExtendedPhotoID{
					ID: "id",
				},
				Path: "to/file",
				PhotoMeta: PhotoMeta{
					Format:    domain.MustFormatForExt("jpg"),
					Location:  gps.MustNewCoordinates(12, 34),
					DateTaken: time.Now(),
				},
				Hash: BinaryHash("1234"),
			},
			JSON: `{
  "id": "id",
  "schema": 6,
  "path": "to/file",
  "format": "jpg",
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
		photo        PhotoMeta
		expectedPath string
		expectedID   string
	}{
		{
			// Reference file
			PhotoMeta{
				Name:        "/some/path/myfile.jpg",
				DateTaken:   at("2015", "02", "24"),
				Location:    somewhere(),
				Format:      domain.MustFormatForExt("jpg"),
				Orientation: 1,
			},
			"2015/02/24",
			// Expect some ID
			"f84a7e9da3f191349ccc603a3e02dba3",
		},
		{
			// Same file name, different path
			PhotoMeta{
				Name:        "/my/other/path/myfile.jpg",
				DateTaken:   at("2015", "02", "24"),
				Location:    somewhere(),
				Format:      domain.MustFormatForExt("jpg"),
				Orientation: 1,
			},
			"2015/02/24",
			// Different ID
			"e73a6689131af5ef009c21b51d137507",
		},
		{
			// Same base name, different extension
			PhotoMeta{
				Name:        "/some/path/myfile.mov",
				DateTaken:   at("2015", "02", "24"),
				Location:    somewhere(),
				Format:      domain.MustFormatForExt("mov"),
				Orientation: 1,
			},
			"2015/02/24",
			// Different ID
			"668fe8834f293464e585ad805cdac9a4",
		},
		{
			// No name given
			PhotoMeta{
				Name:        "",
				DateTaken:   at("2015", "02", "24"),
				Location:    somewhere(),
				Format:      domain.MustFormatForExt("jpg"),
				Orientation: 1,
			},
			"2015/02/24",
			// Different ID
			"a6472197091bf6de095570b06db3e2f3",
		},
		{
			// No name given second time, same parameters
			PhotoMeta{
				Name:        "",
				DateTaken:   at("2015", "02", "24"),
				Location:    somewhere(),
				Format:      domain.MustFormatForExt("jpg"),
				Orientation: 1,
			},
			"2015/02/24",
			// Expect different ID both to referance and previous case
			"df6b6ca195059ad5733bb708178209c5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.photo.Name, func(t *testing.T) {
			actualPath, actualName, id := canonicalizeFilename(tt.photo)
			ext := filepath.Ext(actualName)[1:]
			expectedName := fmt.Sprintf("%s.%s", tt.expectedID, tt.photo.Format.ID())
			assertEquals(t, "extension", tt.photo.Format.ID(), ext)
			assertEquals(t, "id", tt.expectedID, string(id))
			assertEquals(t, "path", tt.expectedPath, actualPath)
			assertEquals(t, "filename", expectedName, actualName)
		})
	}
}

func assertEquals(t *testing.T, name, expected, actual string) {
	if expected != actual {
		t.Errorf("Bad value for '%s': expected '%s', got '%s'", name, expected, actual)
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
