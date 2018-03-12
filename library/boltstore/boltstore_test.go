package boltstore

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
)

type TestFunc func(*testing.T, *BoltStore)

func TestNewStore(t *testing.T) {
	runTestWithStore(t, func(t *testing.T, db *BoltStore) {
		if _, err := os.Stat("/tmp/photos.db"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAddThenFindAll(t *testing.T) {
	runTestWithStore(t, func(t *testing.T, db *BoltStore) {
		photo := randomPhoto()
		if err := db.Add(photo); err != nil {
			t.Fatalf("Failed to add photo: %s", err)
		}
		found := db.FindAll()
		if len(found) != 1 {
			t.Fatalf("Bad number of photos returned, expected %d, got %d", 1, len(found))
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
		idK := sortableId(tK, v.filename)
		idKm1 := sortableId(tKm1, data[k].filename)
		if bytes.Compare(idK, idKm1) <= 0 {
			t.Errorf("Bad byte order, id[%d] is lower than id[%d] (%s <= %s)", k+1, k, tK, tKm1)
		}
	}
}

func runTestWithStore(t *testing.T, test TestFunc) {
	deleteIfExists(t, "/tmp/photos.db")
	defer deleteIfExists(t, "/tmp/photos.db")
	func(t *testing.T) {
		db, err := NewBoltStore("/tmp", "photos.db")
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		switch boltStore := db.(type) {
		case *BoltStore:
			test(t, boltStore)
		default:
			t.Fatal("Bad type: %s", boltStore)
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

func randomPhoto() *library.LibraryPhoto {
	var p library.LibraryPhoto
	data := []byte(`{
		"path" : "2018/02/23",
		"id": "123456789",
		"format": "jpg",
		"date": "2018-02-23T13:43:12Z",
		"gps": {
		   "long": 47.123445,
		   "lat": 45.12313
		}}`)
	if err := json.Unmarshal(data, &p); err != nil {
		panic(err)
	}
	return &p
}
