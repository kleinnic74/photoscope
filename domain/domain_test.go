package domain_test

import (
	"testing"
	"time"

	photos "bitbucket.org/kleinnic74/photos/domain"
)

type PhotoData struct {
	Path   string
	Id     string
	Format string
	Date   string
}

func TestNewPhoto(t *testing.T) {
	var data = []PhotoData{
		{"testdata/orientation/portrait_3.jpg", "portrait_3", "jpg", "2018-02-24T15:19:27+01:00"},
		{"testdata/Canon_40D.jpg", "Canon_40D", "jpg", "2008-05-30T15:56:01+02:00"},
	}
	for _, p := range data {
		act, err := photos.NewPhoto(p.Path)
		if err != nil {
			t.Fatalf("Failed to load image %s: %s", p.Path, err)
		}
		assertExpected(t, &p, act)
	}
}

func assertExpected(t *testing.T, expected *PhotoData, actual photos.Photo) {
	if expected.Format != actual.Format().Id {
		t.Errorf("Bad value for Filename: got %s, expected %s", actual.Format().Id, expected.Format)
	}
	actualDateTaken := actual.DateTaken().Format(time.RFC3339)
	if expected.Date != actualDateTaken {
		t.Errorf("Bad value for DateTaken: got %s, expected %s", actualDateTaken, expected.Date)
	}
	if expected.Id != actual.Id() {
		t.Errorf("Bad value for Id: got %s, expected %s", actual.Id(), expected.Id)
	}
}

func at(t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	if err != nil {
		panic(err)
	}
	return ts.UTC()
}
