package library

import (
	"context"
	"time"

	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type DateIndex interface {
	Keys(context.Context) (Timeline, error)
	Add(context.Context, *Photo) error
	FindRange(context.Context, time.Time, time.Time) ([]PhotoID, error)
}

type GeoIndex interface {
	Has(context.Context, PhotoID) bool
	Get(context.Context, PhotoID) (*gps.Address, bool, error)
	Update(context.Context, PhotoID, *gps.Address) error
}
