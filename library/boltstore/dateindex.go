package boltstore

import (
	"context"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const DateIndexVersion = library.Version(1)

// DateIndex indexes photos by date
type DateIndex struct {
	db *bolt.DB
}

const (
	datesBucketName = "_years"
	dateFormat      = "2006-01-02"
)

var (
	datesBucket = []byte(datesBucketName)
)

// NewDateIndex returns a DateIndex stored in the given BoltDB. If the needed buckets
// do not exist, they will be created
func NewDateIndex(db *bolt.DB) (*DateIndex, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(datesBucket)
		return err
	}); err != nil {
		return nil, err
	}
	return &DateIndex{db: db}, nil
}

// Add will add the given photo to this date index based on its taken time
func (d *DateIndex) Add(ctx context.Context, photo *library.Photo) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		log, _ := logging.FromWithNameAndFields(ctx, "boltdateindex")
		b := tx.Bucket(datesBucket)
		key := d.dayKey(photo.DateTaken)
		dayBucket, err := b.CreateBucketIfNotExists([]byte(key))
		if err != nil {
			log.Warn("Failed to create sub-bucket", zap.String("bucket", key), zap.Error(err))
			return err
		}
		return dayBucket.Put([]byte(photo.ID), []byte(photo.ID))
	})
}

// FindRange returns all photos in the given date range
func (d *DateIndex) FindRangePaged(ctx context.Context, from, to time.Time, start, maxCount int) (ids []library.PhotoID, hasMore bool, err error) {
	from, to = startOfDay(from), endOfDay(to)
	err = d.db.View(func(tx *bolt.Tx) error {
		var count int
		var index int
		b := tx.Bucket(datesBucket)
		for t := from; t.Before(to); t = t.Add(time.Hour * 24) {
			key := d.dayKey(t)
			dayBucket := b.Bucket([]byte(key))
			if dayBucket == nil {
				continue
			}
			c := dayBucket.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				if index < start {
					index++
					continue
				}
				if count == maxCount {
					hasMore = true
					return nil
				}
				ids = append(ids, library.PhotoID(k))
				count++
				index++
			}
		}
		return nil
	})
	return
}

// Keys returns the timeline of indeed photos
func (d *DateIndex) Keys(context.Context) (timeline library.Timeline, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(datesBucket)
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			day, err := time.Parse("2006-01-02", string(k))
			if err != nil {
				return err
			}
			timeline.Add(day, string(k))
		}
		return nil
	})
	return
}

// FindDates returns the days for which photos have been indexed
func (d *DateIndex) FindDates(ctx context.Context) (dates []time.Time, err error) {
	log := logging.From(ctx)
	err = d.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(datesBucket).Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			t, err := time.Parse(dateFormat, string(k))
			if err != nil {
				log.Warn("Bad date in index", zap.String("key", string(k)), zap.Error(err))
				continue
			}
			dates = append(dates, t)
		}
		return nil
	})
	return
}

func (d *DateIndex) dayKey(t time.Time) string {
	return t.Format(dateFormat)
}

func startOfDay(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func endOfDay(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 999999999, d.Location())
}
