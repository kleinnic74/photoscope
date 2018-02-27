// format_test.go
package domain_test

import (
	"io"
	"os"
	"testing"

	"bitbucket.org/kleinnic74/photos/domain"
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
		if actual.Id != i.expExt {
			t.Errorf("Bad extension for %s: Expectetd %s, got %s", i.t, i.expExt, actual.Id)
		}
	}
}

func TestJpegFormat(t *testing.T) {
	r := mustOpenFile(t, "testdata/Canon_40D.jpg")
	f, err := domain.FormatOf(r)
	if err != nil {
		t.Fatal(err)
	}
	if f.Id != "jpg" {
		t.Errorf("Bad format, expected %s, got %s", "jpg", f)
	}
}

func mustOpenFile(t *testing.T, path string) io.Reader {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Could not open test file %s: %s", path, err)
	}
	return f
}
