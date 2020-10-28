// Package boltstore is an implementation of a library meta-data index
// using BoltDB for storing data persistently
package boltstore

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/library"

	"github.com/boltdb/bolt"
	"github.com/reusee/mmh3"
)

var (
	photosBucket = []byte("photos")
	hashBucket   = []byte("photoHashes")
	idMapBucket  = []byte("idmap")
)

// BoltStore uses BoltDB as the storage implementation to store data about photos
type BoltStore struct {
	db      *bolt.DB
	indexes []interface{}
}

// NewBoltStore creates a new BoltStore at the given location with the given name
func NewBoltStore(db *bolt.DB) (lib library.ClosableStore, err error) {
	lib = &BoltStore{
		db: db,
	}
	if !db.IsReadOnly() {
		if err = createBucket(db, photosBucket); err != nil {
			return
		}
		if err = createBucket(db, idMapBucket); err != nil {
			return
		}
		if err = createBucket(db, hashBucket); err != nil {
			return
		}
	}
	return
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

func deleteBuckets(db *bolt.DB, names ...string) error {
	return db.Update(func(tx *bolt.Tx) error {
		for _, name := range names {
			b := tx.Bucket([]byte(name))
			if b == nil {
				continue
			}
			if err := tx.DeleteBucket([]byte(name)); err != nil {
				return err
			}
		}
		return nil
	})
}

// Close closes this store
func (store *BoltStore) Close() {
}

// Exists checks if a photo with the given id on the given date exists in this store
func (store *BoltStore) Exists(hash library.BinaryHash) (other library.PhotoID, exists bool) {
	store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(hashBucket)
		id := b.Get(hash.Bytes())
		exists = id != nil
		other = library.PhotoID(id)
		return nil
	})
	return
}

// Add adds the given photo to this store
func (store *BoltStore) Add(p *library.Photo) error {
	id := sortableID(p.DateTaken, string(p.ID))
	encoded, err := json.Marshal(p)
	if err != nil {
		log.Printf("Error: failed to encode photo: %s", err)
		return err
	}
	return store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosBucket)
		if existing := b.Get(id); existing != nil {
			return library.PhotoAlreadyExists(p.ID)
		}
		err = b.Put(id, encoded)
		if err != nil {
			return err
		}
		b = tx.Bucket(idMapBucket)
		err = b.Put([]byte(p.ID), id)
		if err != nil {
			return err
		}
		if p.HasHash() {
			b = tx.Bucket(hashBucket)
			err = b.Put([]byte(p.Hash), []byte(p.ID))
		}
		return err
	})
}

func (store *BoltStore) Update(p *library.Photo) error {
	encoded, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return store.db.Update(func(tx *bolt.Tx) error {
		internalID := tx.Bucket(idMapBucket).Get([]byte(p.ID))
		if internalID == nil {
			return library.NotFound(p.ID)
		}
		b := tx.Bucket(photosBucket)
		if err := b.Put(internalID, encoded); err != nil {
			return err
		}
		if !p.HasHash() {
			return nil
		}
		b = tx.Bucket(hashBucket)
		return b.Put([]byte(p.Hash), []byte(p.ID))
	})
}

// FindAll returns all photos in this store
func (store *BoltStore) FindAll(order consts.SortOrder) ([]*library.Photo, error) {
	photos, _, err := store.findRange(func(c Cursor) Cursor {
		return c
	}, order)
	return photos, err
}

//FindAllPaged returns at most max photos from the store starting at photo index start
func (store *BoltStore) FindAllPaged(start, max int, order consts.SortOrder) ([]*library.Photo, bool, error) {
	return store.findRange(func(c Cursor) Cursor {
		return c.Skip(uint(start)).Limit(uint(max))
	}, order)
}

func (store *BoltStore) findRange(f func(Cursor) Cursor, order consts.SortOrder) ([]*library.Photo, bool, error) {
	var found = make([]*library.Photo, 0)
	var hasMore bool

	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosBucket)
		c := f(newCursor(b.Cursor(), order))
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var photo library.Photo
			if err := json.Unmarshal(v, &photo); err != nil {
				log.Printf("Error: could not unmarshal photo: %s", err)
				return err
			}
			found = append(found, &photo)
		}
		hasMore = c.HasMore()
		return nil
	})
	if err != nil {
		log.Printf("Could not read photos: %s", err)
	}
	return found, hasMore, err
}

// Find returns all photos in this library between the given time instants
func (store *BoltStore) Find(start, end time.Time, order consts.SortOrder) ([]*library.Photo, error) {
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
	return found, err
}

// Get returns the photo with the given id
func (store *BoltStore) Get(id library.PhotoID) (*library.Photo, error) {
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
