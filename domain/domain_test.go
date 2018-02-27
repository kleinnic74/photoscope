package domain_test

import (
	"testing"
	"time"

	photos "bitbucket.org/kleinnic74/photos/domain"
)

type PhotoData struct {
	path string
	name string
	date string
}

func TestNewPhoto(t *testing.T) {
	var data = []PhotoData{
		{"testdata/orientation/portrait_3.jpg", "portrait_3.jpg", "2018-02-24T15:19:27+01:00"},
		{"testdata/Canon_40D.jpg", "Canon_40D.jpg", "2008-05-30T15:56:01+02:00"},
	}
	for _, p := range data {
		act, err := photos.NewPhoto(p.path)
		if err != nil {
			t.Fatalf("Failed to load image %s: %s", p.path, err)
		}
		assertExpected(t, &p, act)
	}
}

func assertExpected(t *testing.T, expected *PhotoData, actual *photos.Photo) {
	if expected.path != actual.Path {
		t.Errorf("Bad value for Path: got %s, expected %s", actual.Path, expected.path)
	}
	if expected.name != actual.Filename {
		t.Errorf("Bad value for Filename: got %s, expected %s", actual.Filename, expected.name)
	}
	actualDateTaken := actual.DateTaken.Format(time.RFC3339)
	if expected.date != actualDateTaken {
		t.Errorf("Bad value for DateTaken: got %s, expected %s", actualDateTaken, expected.date)
	}
}
