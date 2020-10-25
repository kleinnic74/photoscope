// format_test.go
package domain_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"bitbucket.org/kleinnic74/photos/domain"
	"github.com/stretchr/testify/assert"
)

func TestFormatById(t *testing.T) {
	test := []struct {
		t       string
		expExt  string
		expMime string
	}{
		{"jpg", "jpg", "image/jpeg"},
		{"mov", "mov", "video/quicktime"},
	}
	for _, i := range test {
		actual, found := domain.FormatForExt(i.t)
		if !found {
			t.Errorf("Expected format for ext %s, but returned nothing", i.t)
		}
		if actual.ID() != i.expExt {
			t.Errorf("Bad extension for %s: Expectetd %s, got %s", i.t, i.expExt, actual.ID())
		}
	}
}

func TestJpegFormat(t *testing.T) {
	r := mustOpenFile(t, "testdata/Canon_40D.jpg")
	defer r.Close()
	f, err := domain.FormatOf(r)
	if err != nil {
		t.Fatal(err)
	}
	if f.ID() != "jpg" {
		t.Errorf("Bad format, expected %s, got %s", "jpg", f.ID())
	}
}

func TestUnmarshalJSON(t *testing.T) {
	data := []struct {
		ext            string
		expectedFormat domain.Format
	}{
		{`"jpg"`, domain.JPEG},
		{`"JPG"`, domain.JPEG},
		{`"MOV"`, domain.MOV},
		{`"mov"`, domain.MOV},
		{`"bad"`, domain.UnknownFormat},
	}
	for _, d := range data {
		var f domain.FormatSpec
		if err := json.Unmarshal([]byte(d.ext), &f); err != nil {
			t.Fatalf("Error while JSON decoding: %s", err)
		}
		t.Logf("FormatSpec: %s", string(f))
		assert.Equal(t, d.expectedFormat.Mime(), f.Mime())
	}
}

func mustOpenFile(t *testing.T, path string) io.ReadCloser {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Could not open test file %s: %s", path, err)
	}
	return f
}
