package boltstore

import (
	"context"

	"bitbucket.org/kleinnic74/photos/index"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

func resetBuckets(db *bolt.DB, buckets ...[]byte) index.StructuralMigration {
	return index.StructuralMigrationFunc(func(ctx context.Context) (bool, error) {
		log, ctx := logging.SubFrom(ctx, "resetBuckets")
		log.Debug("Resetting buckets...")
		tx, err := db.Begin(true)
		if err != nil {
			return false, err
		}
		for _, b := range buckets {
			log.Info("Re-creating bucket", zap.String("bucket", string(b)))
			if tx.Bucket(b) != nil {
				if err := tx.DeleteBucket(b); err != nil {
					return false, err
				}
				log.Info("Deleted old bucket", zap.String("bucket", string(b)))
			}
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return false, err
			}
			log.Info("Re-created bucket", zap.String("bucket", string(b)))
		}
		log.Debug("Done")
		return true, tx.Commit()
	})
}
