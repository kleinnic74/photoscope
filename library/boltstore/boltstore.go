// Package boltstore is an implementation of a library meta-data index
// using BoltDB for storing data persistently
package boltstore

import (
	"bytes"
	"encoding/json"
	"fmt"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/library"

	bolt "go.etcd.io/bbolt"
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
	// Sanity check
	if p.ID == "" {
		panic(fmt.Errorf("Photo %v has no ID", p))
	}
	if len(p.SortID) == 0 {
		panic(fmt.Errorf("Photo %s has no SortID", p.ID))
	}

	encoded, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosBucket)
		if existing := b.Get(p.SortID); existing != nil {
			return library.PhotoAlreadyExists(p.ID)
		}
		err = b.Put(p.SortID, encoded)
		if err != nil {
			return err
		}
		b = tx.Bucket(idMapBucket)
		err = b.Put([]byte(p.ID), p.SortID)
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
	// Sanity check
	if p.ID == "" {
		panic(fmt.Errorf("Photo %v has no ID", p))
	}
	if len(p.SortID) == 0 {
		panic(fmt.Errorf("Photo %s has no SortID", p.ID))
	}

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
				return err
			}
			found = append(found, &photo)
		}
		hasMore = c.HasMore()
		return nil
	})
	return found, hasMore, err
}

// Find returns all photos in this library between the given time instants
func (store *BoltStore) Find(start, end library.OrderedID, order consts.SortOrder) ([]*library.Photo, error) {
	var found = make([]*library.Photo, 0)
	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosBucket)
		c := b.Cursor()

		for k, v := c.Seek(start); k != nil && bytes.Compare(k, end) <= 0; k, v = c.Next() {
			var photo library.Photo
			if err := json.Unmarshal(v, &photo); err != nil {
				return err
			}
			found = append(found, &photo)
		}
		return nil
	})
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
				return err
			}
			found = &photo
			return nil
		}
		return library.NotFound(id)
	})
}
