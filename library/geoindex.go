package library

import (
	"context"

	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type GeoIndex interface {
	Has(context.Context, string) bool
	Get(context.Context, string) (*gps.Address, bool, error)
	Update(context.Context, string, *gps.Address) error
}
