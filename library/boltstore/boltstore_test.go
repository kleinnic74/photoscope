package boltstore

import (
	"os"
	"path/filepath"
	"testing"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/library"
	bolt "go.etcd.io/bbolt"
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
		found, _ := db.FindAll(consts.Ascending)
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
		found, err := db.Get(photo.ID)
		if err != nil {
			t.Fatalf("Should have found a photo with id %s", photo.ID)
		}
		if found == nil {
			t.Fatalf("Returned nil and nil error, error should have been NotFound")
		}
		assertPhotosAreEqual(t, photo, found)
	})
}

func BenchmarkAdd(b *testing.B) {
	// Initialize store
	dbFile := filepath.Join(dbpath, dbfile)
	boltDb, err := bolt.Open(dbFile, 0644, nil)
	if err != nil {
		b.Fatal(err)
	}
	db, err := NewBoltStore(boltDb)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		db.Close()
		os.Remove(dbFile)
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

type BoltDBTestFunc func(t *testing.T, db *bolt.DB)

func runTestWithBoltDB(t *testing.T, test BoltDBTestFunc) {
	fullpath := filepath.Join(dbpath, dbfile)
	deleteIfExists(t, fullpath)
	defer deleteIfExists(t, fullpath)
	os.Mkdir(dbpath, 0755)
	func(t *testing.T) {
		boltDB, err := bolt.Open(fullpath, 0644, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer boltDB.Close()
		test(t, boltDB)
	}(t)

}

func runTestWithStore(t *testing.T, test TestFunc) {
	runTestWithBoltDB(t, func(t *testing.T, boltDB *bolt.DB) {
		db, err := NewBoltStore(boltDB)
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
	})
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
		t.Errorf("Both should be nil but are not: p1=%v, p2=%v", p1, p2)
	}
	if p1.ID != p2.ID {
		t.Errorf("Different Id()s: p1: %s, p2: %s", p1.ID, p2.ID)
	}
}
