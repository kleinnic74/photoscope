package boltstore

import (
	"context"
	"encoding/json"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

var (
	eventsBucket        = []byte("_events")
	photosByEventBucket = []byte("_photosByEvent")
)

type EventID string

type Event struct {
	ID   EventID   `json:"id"`
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

func (index *EventIndex) AddPhotosToEvent(ctx context.Context, e Event, photos []library.ExtendedPhotoID) error {
	log, ctx := logging.FromWithNameAndFields(ctx, "eventindex", zap.String("event", string(e.ID)))
	encoded, err := json.Marshal(&e)
	if err != nil {
		return err
	}
	err = index.db.Update(func(tx *bolt.Tx) error {
		log.Info("Updating event", zap.Int("nbPhotos", len(photos)))
		if err := tx.Bucket(eventsBucket).Put([]byte(e.ID), encoded); err != nil {
			return err
		}
		b, err := tx.Bucket(photosByEventBucket).CreateBucketIfNotExists([]byte(e.ID))
		if err != nil {
			return err
		}
		for _, p := range photos {
			if err := b.Put([]byte(p.SortID), []byte(p.ID)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Warn("Event update failed", zap.Error(err))
	}
	return err
}

func (index *EventIndex) FindPaged(ctx context.Context, start, maxCount int) (events []Event, hasMore bool, err error) {
	err = index.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(eventsBucket).Cursor()
		var (
			i    int
			k, v []byte
		)
		for k, v = c.First(); k != nil && i < start+maxCount; k, v = c.Next() {
			if i < start {
				i++
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

func (index *EventIndex) FindPhotosPaged(ctx context.Context, eventID string, start, max int) (photos []library.PhotoID, hasMore bool, err error) {
	err = index.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosByEventBucket).Bucket([]byte(eventID))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		var k, v []byte
		var i int
		for k, v = c.First(); k != nil && i < start+max; k, v = c.Next() {
			if i < start {
				i++
				continue
			}
			photos = append(photos, library.PhotoID(v))
			i++
		}
		hasMore = k != nil
		return nil
	})
	return
}
