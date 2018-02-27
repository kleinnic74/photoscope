package library

import (
	"testing"

	"bitbucket.org/kleinnic74/photos/domain"
)

func TestCanonicalizePhoto(t *testing.T) {
	var tests = []struct {
		photo        domain.Photo
		expectedPath string
		expectedName string
	}{
		{domain.Photo{}, "2015/02/24", ""},
	}
	for _, tt := range tests {
		actualPath, actualName := canonicalizeFilename(&tt.photo)
		if actualPath != tt.expectedPath {
			t.Errorf("%s: expected path: %s, got path: %s", &tt.photo, tt.expectedPath, actualPath)
		}
	}
}
