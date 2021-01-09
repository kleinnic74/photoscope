package library

import (
	"context"

	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type Locations struct {
	Countries []*CountryAndPlaces `json:"countries"`
}

type CountryAndPlaces struct {
	gps.Country
	Places []*gps.Address `json:"places"`
}

type GeoIndex interface {
	MigrateStructure(context.Context, Version) (Version, bool, error)

	Has(context.Context, PhotoID) bool
	Get(context.Context, PhotoID) (*gps.Address, bool, error)
	Update(context.Context, ExtendedPhotoID, *gps.Address) error

	Locations(context.Context) (*Locations, error)
	FindByPlacePaged(context.Context, gps.PlaceID, int, int) ([]PhotoID, bool, error)
	FindByCountryPaged(context.Context, gps.CountryID, int, int) ([]PhotoID, bool, error)
}
