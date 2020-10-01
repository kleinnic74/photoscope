package library

import (
	"context"
	"time"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

type DateIndex interface {
	Add(context.Context, *Photo) error
	FindRange(context.Context, time.Time, time.Time) ([]string, error)
}

type GeoIndex interface {
	Has(context.Context, string) bool
	Get(context.Context, string) (*gps.Address, bool, error)
	Update(context.Context, string, *gps.Address) error
}

func IndexByDate(index DateIndex) NewPhotoCallback {
	return func(ctx context.Context, photo *Photo) {
		if err := index.Add(ctx, photo); err != nil {
			log, _ := logging.FromWithNameAndFields(ctx, "dateindex")
			log.Warn("Failed to add photo", zap.Error(err))
		}
	}
}
