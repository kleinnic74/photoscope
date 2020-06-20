// Package boltstore is an implementation of a library meta-data index
// using BoltDB for storing data persistently
package boltstore

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/library"

	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/reusee/mmh3"
)

var (
	photosBucket = []byte("photos")
	idMapBucket  = []byte("idmap")
)

// BoltStore uses BoltDB as the storage implementation to store data about photos
type BoltStore struct {
	db *bolt.DB
}

// NewBoltStore creates a new BoltStore at the given location with the given name
func NewBoltStore(basedir string, name string) (library.ClosableStore, error) {
	db, err := bolt.Open(filepath.Join(basedir, name), 0600, nil)
	if err != nil {
		return nil, err
	}
	if err = createBucket(db, photosBucket); err != nil {
		db.Close()
		return nil, err
	}
	if err = createBucket(db, idMapBucket); err != nil {
		db.Close()
		return nil, err
	}
	return &BoltStore{
		db: db,
	}, nil
}

func createBucket(db *bolt.DB, name []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(name)
		if err != nil {
			return err
		}
		return nil
	})
}

// Close closes this store
func (store *BoltStore) Close() {
	store.db.Close()
}

// Exists checks if a photo with the given id on the given date exists in this store
func (store *BoltStore) Exists(dateTaken time.Time, id string) bool {
	key := sortableID(dateTaken, id)
	var exists bool
	store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosBucket)
		exists = b.Get(key) != nil
		return nil
	})
	return exists
}

// Add adds the given photo to this store
func (store *BoltStore) Add(p *library.Photo) error {
	id := sortableID(p.DateTaken(), p.ID())
	encoded, err := json.Marshal(p)
	if err != nil {
		log.Printf("Error: failed to encode photo: %s", err)
		return err
	}
	return store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosBucket)
		if existing := b.Get(id); existing != nil {
			return library.PhotoAlreadyExists(p.ID())
		}
		err = b.Put(id, encoded)
		if err != nil {
			return err
		}
		b = tx.Bucket(idMapBucket)
		err = b.Put([]byte(p.ID()), id)
		return err
	})
}

// FindAll returns all photos in this store
func (store *BoltStore) FindAll() []*library.Photo {
	return store.findRange(func(c Cursor) Cursor {
		return c
	})
}

//FindAllPaged returns at most max photos from the store starting at photo index start
func (store *BoltStore) FindAllPaged(start, max uint) []*library.Photo {
	return store.findRange(func(c Cursor) Cursor {
		return c.Skip(start).Limit(max)
	})
}

func (store *BoltStore) findRange(f func(Cursor) Cursor) []*library.Photo {
	var found = make([]*library.Photo, 0)

	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosBucket)
		c := f(newForwardCursor(b.Cursor()))
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var photo library.Photo
			if err := json.Unmarshal(v, &photo); err != nil {
				log.Printf("Error: could not unmarshal photo: %s", err)
				return err
			}
			found = append(found, &photo)
		}
		return nil
	})
	if err != nil {
		log.Printf("Could not read photos: %s", err)
	}
	return found
}

// Find returns all photos in this library between the given time instants
func (store *BoltStore) Find(start, end time.Time) []*library.Photo {
	var found = make([]*library.Photo, 0)
	min, max := boundaryIDs(start, end)
	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosBucket)
		c := b.Cursor()

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			var photo library.Photo
			if err := json.Unmarshal(v, &photo); err != nil {
				log.Printf("Error: could not unmarshal photo: %s", err)
				return err
			}
			found = append(found, &photo)
		}
		return nil
	})
	if err != nil {
		log.Printf("Could not read photos: %s", err)
	}
	return found
}

// Get returns the photo with the given id
func (store *BoltStore) Get(id string) (*library.Photo, error) {
	var found *library.Photo
	return found, store.db.View(func(tx *bolt.Tx) error {
		internalID := tx.Bucket(idMapBucket).Get([]byte(id))
		if internalID == nil {
			return library.NotFound(id)
		}
		var photo library.Photo
		if data := tx.Bucket(photosBucket).Get(internalID); data != nil {
			if err := json.Unmarshal(data, &photo); err != nil {
				log.Printf("Error: could not unmarshal photo: %s", err)
				return err
			}
			found = &photo
			return nil
		}
		return library.NotFound(id)
	})
}

func sortableID(ts time.Time, filename string) []byte {
	var id bytes.Buffer
	id.Write([]byte(ts.UTC().Format(time.RFC3339)))
	h := mmh3.New32()
	h.Write([]byte(strings.ToLower(filename)))
	id.Write(h.Sum(nil))
	return id.Bytes()
}

func boundaryIDs(begin, end time.Time) (low, high []byte) {
	var lbuf bytes.Buffer
	lbuf.Write([]byte(begin.UTC().Format(time.RFC3339)))
	lbuf.Write([]byte{0, 0, 0, 0})
	low = lbuf.Bytes()
	var hbuf bytes.Buffer
	hbuf.Write([]byte(end.UTC().Format(time.RFC3339)))
	hbuf.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	high = lbuf.Bytes()
	return
}
