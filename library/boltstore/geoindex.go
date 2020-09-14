package boltstore

import (
	"context"
	"encoding/json"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

type boltGeoIndex struct {
	db *bolt.DB
}

var placesBucket = []byte("photoplaces")

func NewBoltGeoIndex(db *bolt.DB) (library.GeoIndex, error) {
	if err := createBucket(db, placesBucket); err != nil {
		return nil, err
	}
	return &boltGeoIndex{
		db: db,
	}, nil
}

func (index *boltGeoIndex) Has(ctx context.Context, id string) (exists bool) {
	logger, ctx := logging.FromWithNameAndFields(ctx, "geoStore")
	err := index.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(placesBucket)
		exists = b.Get([]byte(id)) != nil
		return nil
	})
	if err != nil {
		logger.Warn("Bolt error", zap.String("photo", id), zap.Error(err))
	}
	return
}

func (index *boltGeoIndex) Get(ctx context.Context, id string) (address *gps.Address, found bool, err error) {
	logger, ctx := logging.FromWithNameAndFields(ctx, "geoStore")
	if err := index.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(placesBucket)
		data := b.Get([]byte(id))
		found = data != nil
		if !found {
			return nil
		}
		address = new(gps.Address)
		return json.Unmarshal(data, address)
	}); err != nil {
		logger.Warn("Bolt error", zap.String("photo", id), zap.Error(err))
		return nil, false, err
	}
	return
}

func (index *boltGeoIndex) Update(ctx context.Context, id string, address *gps.Address) error {
	if address == nil {
		return nil
	}
	encoded, err := json.Marshal(address)
	if err != nil {
		return err
	}
	return index.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(placesBucket)
		return b.Put([]byte(id), encoded)
	})
}
