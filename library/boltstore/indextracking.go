package boltstore

import (
	"encoding/json"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/library/index"
	"github.com/boltdb/bolt"
)

var (
	indexBucket = []byte("_indextracker")
)

type indexTracker struct {
	db      *bolt.DB
	indexes map[index.Name]struct{}
}

// NewIndexTracker returns a new index tracker using the given BoltDB. The needed
// buckets are created if not yet available.
func NewIndexTracker(db *bolt.DB) (index.Tracker, error) {
	if err := db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists(indexBucket)
		return
	}); err != nil {
		return nil, err
	}
	return &indexTracker{db: db, indexes: make(map[index.Name]struct{})}, nil
}

func (tracker *indexTracker) RegisterIndex(index index.Name) {
	tracker.indexes[index] = struct{}{}
}

func (tracker *indexTracker) Update(name index.Name, id library.PhotoID, err error) error {
	var status index.Status
	if err != nil {
		status = index.ErrorOnIndex
	} else {
		status = index.Indexed
	}
	return tracker.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(indexBucket)
		state := index.NewState()
		if v := b.Get([]byte(id)); v != nil {
			if err := json.Unmarshal(v, &state); err != nil {
				return err
			}
		}
		state.Set(name, status)
		stateBytes, err := json.Marshal(state)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), stateBytes)
	})
}

func (tracker *indexTracker) Get(id library.PhotoID) (index.State, bool, error) {
	var found bool
	state := index.NewState()
	err := tracker.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(indexBucket).Get([]byte(id))
		if v == nil {
			return nil
		}
		found = true
		return json.Unmarshal(v, &state)
	})
	return state, found, err
}

func (tracker *indexTracker) GetMissingIndexes(id library.PhotoID) (missing []index.Name, err error) {
	err = tracker.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(indexBucket).Get([]byte(id))
		state := index.NewState()
		if v != nil {
			if err := json.Unmarshal(v, &state); err != nil {
				return err
			}
		}
		for k := range tracker.indexes {
			if state.StatusFor(k) == index.NotIndexed {
				missing = append(missing, k)
			}
		}
		return err
	})
	return
}
