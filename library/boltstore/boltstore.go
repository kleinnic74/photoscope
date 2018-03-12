// boltstore.go
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
	PhotosBucket = []byte("photos")
)

type BoltStore struct {
	db *bolt.DB
}

func NewBoltStore(basedir string, name string) (library.ClosableStore, error) {
	db, err := bolt.Open(filepath.Join(basedir, name), 0600, nil)
	if err != nil {
		return nil, err
	}
	if err = createBucket(db, PhotosBucket); err != nil {
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

func (store *BoltStore) Close() {
	store.db.Close()
}

func (story *BoltStore) Exists(dateTaken time.Time, id string) bool {
	key := sortableId(dateTaken, id)
	var exists bool
	story.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(PhotosBucket)
		exists = b.Get(key) != nil
		return nil
	})
	return exists
}

func (store *BoltStore) Add(p *library.LibraryPhoto) error {
	id := sortableId(p.DateTaken(), p.Id())
	return store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(PhotosBucket)
		if existing := b.Get(id); existing != nil {
			return library.PhotoAlreadyExists
		}
		encoded, err := json.Marshal(p)
		if err != nil {
			log.Printf("Error: failed to encode photo: %s", err)
			return err
		}
		err = b.Put(id, encoded)
		return err
	})
}

func (store *BoltStore) FindAll() []*library.LibraryPhoto {
	var found []*library.LibraryPhoto = make([]*library.LibraryPhoto, 0)

	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(PhotosBucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var photo library.LibraryPhoto
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

func (store *BoltStore) Find(start, end time.Time) []*library.LibraryPhoto {
	var found []*library.LibraryPhoto = make([]*library.LibraryPhoto, 0)
	min, max := boundaryIds(start, end)
	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(PhotosBucket)
		c := b.Cursor()

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			var photo library.LibraryPhoto
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
	log.Printf("Entries found: %d", len(found))
	return found
}

func sortableId(ts time.Time, filename string) []byte {
	var id bytes.Buffer
	id.Write([]byte(ts.UTC().Format(time.RFC3339)))
	h := mmh3.New32()
	h.Write([]byte(strings.ToLower(filename)))
	id.Write(h.Sum(nil))
	return id.Bytes()
}

func boundaryIds(begin, end time.Time) (low, high []byte) {
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
