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
	Places []*gps.Place `json:"places"`
}

type GeoIndex interface {
	Has(context.Context, PhotoID) bool
	Get(context.Context, PhotoID) (*gps.Address, bool, error)
	Update(context.Context, PhotoID, *gps.Address) error

	Locations(context.Context) (*Locations, error)
	FindByPlacePaged(context.Context, string, string, int, int) ([]PhotoID, bool, error)
	FindByCountryPaged(context.Context, string, int, int) ([]PhotoID, bool, error)
}
