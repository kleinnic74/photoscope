package boltstore

import (
	"errors"
	"testing"

	"bitbucket.org/kleinnic74/photos/index"
	"bitbucket.org/kleinnic74/photos/library"
	"github.com/boltdb/bolt"
)

func TestAddToIndexTracker(t *testing.T) {
	runTestWithBoltDB(t, testAddToIndexTracker)
}

func testAddToIndexTracker(t *testing.T, db *bolt.DB) {
	tracker, err := NewIndexTracker(db)
	if err != nil {
		t.Fatalf("Failed to init index tracker: %s", err)
	}
	tracker.RegisterIndex("geo", library.Version(1))
	data := []struct {
		ID                  library.PhotoID
		err                 error
		updateFirst         bool
		found               bool
		expectedIndexStatus index.Status
	}{
		{"1234", nil, false, false, index.NotIndexed},
		{"1234", nil, true, true, index.Indexed},
		{"1234", errors.New("some error"), true, true, index.ErrorOnIndex},
	}
	for i, d := range data {
		if d.updateFirst {
			if err := tracker.Update("geo", d.ID, d.err); err != nil {
				t.Fatalf("#%d: failed to update index: %s", i, err)
			}
			state, found, err := tracker.Get(d.ID)
			if err != nil {
				t.Fatalf("#%d: failed to retrieve state: %s", i, err)
			}
			if found != d.found {
				t.Fatalf("#%d: bad found indication, expected %t, got %t", i, d.found, found)
			}
			actualStatus := state.StatusFor("geo").Status
			if actualStatus != d.expectedIndexStatus {
				t.Errorf("#%d: bad status for index %s, expected %d, got %d", i, "geo", d.expectedIndexStatus, actualStatus)
			}
		}
	}
}
