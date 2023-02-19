package domain_test

import (
	"os"
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
		{"testdata/orientation/portrait_3.jpg", "portrait_3", "jpg", fileModificationTime("testdata/orientation/portrait_3.jpg")},
		// The following image does not have time zone information, using local zone
		{"testdata/Canon_40D.jpg", "Canon_40D", "jpg", localTime("2008-05-30T15:56:01")},
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
	if expected.Format != actual.Format().ID() {
		t.Errorf("%s: Bad value for Filename: got %s, expected %s", expected.Path, actual.Format().ID(), expected.Format)
	}
	actualDateTaken := actual.DateTaken().Format(time.RFC3339)
	if expected.Date != actualDateTaken {
		t.Errorf("%s: Bad value for DateTaken: got %s, expected %s", expected.Path, actualDateTaken, expected.Date)
	}
	if expected.Id != actual.ID() {
		t.Errorf("%s: Bad value for Id: got %s, expected %s", expected.Path, actual.ID(), expected.Id)
	}
}

func at(t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	if err != nil {
		panic(err)
	}
	return ts.UTC()
}

func fileModificationTime(path string) string {
	info, _ := os.Stat(path)
	return info.ModTime().Format(time.RFC3339)
}

func localTime(ts string) string {
	t, _ := time.ParseInLocation("2006-01-02T15:04:05", ts, time.Local)
	t = t.Local()
	return t.Format("2006-01-02T15:04:05Z07:00")
}
