package boltstore

import (
	"context"
	"encoding/json"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"github.com/boltdb/bolt"
)

var (
	eventsBucket        = []byte("_events")
	photosByEventBucket = []byte("_photosByEvent")
)

type Event struct {
	ID   string    `json:"id"`
	Name string    `json:"name,omitempty"`
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type EventIndex struct {
	db *bolt.DB
}

func NewEventIndex(db *bolt.DB) (*EventIndex, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(eventsBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(photosByEventBucket); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &EventIndex{
		db: db,
	}, nil
}

func (index *EventIndex) Add(ctx context.Context, e Event) error {
	return index.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(eventsBucket)
		v, err := json.Marshal(&e)
		if err != nil {
			return err
		}
		return b.Put([]byte(e.ID), v)
	})
}

func (index *EventIndex) AddPhotosToEvent(ctx context.Context, e Event, photos []library.PhotoID) error {
	return index.db.Update(func(tx *bolt.Tx) error {
		return nil
	})
}

func (index *EventIndex) FindPaged(ctx context.Context, start, maxCount int) (events []Event, hasMore bool, err error) {
	err = index.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(eventsBucket).Cursor()
		var (
			i int
			k []byte
			v []byte
		)
		for k, v = c.First(); k != nil && i < start+maxCount; k, v = c.Next() {
			if i < start {
				continue
			}
			var e Event
			if err := json.Unmarshal(v, &e); err != nil {
				return err
			}
			events = append(events, e)
			i++
		}
		hasMore = k != nil
		return nil
	})
	return
}
