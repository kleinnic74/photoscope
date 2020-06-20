package boltstore

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
)

const (
	dbpath = "test"
	dbfile = "photos.db"
)

type TestFunc func(*testing.T, *BoltStore)

func TestNewStore(t *testing.T) {
	runTestWithStore(t, func(t *testing.T, db *BoltStore) {
		if _, err := os.Stat(filepath.Join(dbpath, dbfile)); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAddThenFindAll(t *testing.T) {
	runTestWithStore(t, func(t *testing.T, db *BoltStore) {
		photo := library.RandomPhoto()
		if err := db.Add(photo); err != nil {
			t.Fatalf("Failed to add photo: %s", err)
		}
		found := db.FindAll()
		if len(found) != 1 {
			t.Fatalf("Bad number of photos returned, expected %d, got %d", 1, len(found))
		}
	})
}

func TestAddThenGet(t *testing.T) {
	runTestWithStore(t, func(t *testing.T, db *BoltStore) {
		photo := library.RandomPhoto()
		if err := db.Add(photo); err != nil {
			t.Fatalf("Failed to add photo: %s", err)
		}
		found, err := db.Get(photo.ID())
		if err != nil {
			t.Fatalf("Should have found a photo with id %s", photo.ID())
		}
		if found == nil {
			t.Fatalf("Returned nil and nil error, error should have been NotFound")
		}
		assertPhotosAreEqual(t, photo, found)
	})
}

func BenchmarkAdd(b *testing.B) {
	// Initialize store
	db, err := NewBoltStore(dbpath, dbfile)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		db.Close()
		os.Remove("/tmp/photos.db")
	}()

	b.Run("Add a photo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := library.RandomPhoto()
			if err := db.Add(p); err != nil {
				b.Errorf("Failed to add photo: %s", err)
			}
		}
	})
}

func TestByteOrderOfId(t *testing.T) {
	var data = []struct {
		ts       string
		filename string
	}{
		{"2016-12-26T17:09:11Z", "cde"},
		{"2016-12-31T17:46:11Z", "cde"},
		{"2017-02-24T15:22:18Z", "abc"},
	}
	for k, v := range data[1:] {
		tK, _ := time.Parse(time.RFC3339, v.ts)
		tKm1, _ := time.Parse(time.RFC3339, data[k].ts)
		if tK.Before(tKm1) {
			t.Fatalf("Bad time stamp order, ts[%d] is before ts[%d]", k+1, k)
		}
		idK := sortableID(tK, v.filename)
		idKm1 := sortableID(tKm1, data[k].filename)
		if bytes.Compare(idK, idKm1) <= 0 {
			t.Errorf("Bad byte order, id[%d] is lower than id[%d] (%s <= %s)", k+1, k, tK, tKm1)
		}
	}
}

func runTestWithStore(t *testing.T, test TestFunc) {
	fullpath := filepath.Join(dbpath, dbfile)
	deleteIfExists(t, fullpath)
	defer deleteIfExists(t, fullpath)
	os.Mkdir(dbpath, 0755)
	func(t *testing.T) {
		db, err := NewBoltStore(dbpath, dbfile)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		switch boltStore := db.(type) {
		case *BoltStore:
			test(t, boltStore)
		default:
			t.Fatalf("Bad type: %s", boltStore)
		}
	}(t)
}

func deleteIfExists(t *testing.T, file string) {
	info, err := os.Stat(file)
	if err != nil {
		return
	}
	if err == nil {
		if err = os.Remove(file); err != nil {
			t.Fatal(err)
		}
	}
	if info.IsDir() {
		t.Fatalf("%s is directory", file)
	}
}

func assertPhotosAreEqual(t *testing.T, p1, p2 *library.Photo) {
	if (p1 == nil || p2 == nil) && p1 != p2 {
		t.Errorf("Both should be nil but are not: p1=%s, p2=%s", p1, p2)
	}
	if p1.ID() != p2.ID() {
		t.Errorf("Different Id()s: p1: %s, p2: %s", p1.ID(), p2.ID())
	}
}
