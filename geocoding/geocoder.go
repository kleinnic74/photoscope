package geocoding

import (
	"context"
	"errors"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/tasks"
	"github.com/codingsince1985/geo-golang"
	"github.com/codingsince1985/geo-golang/openstreetmap"
	"go.uber.org/zap"
)

var UnknownLocation = errors.New("Unknown GPS location")

type Geocoder struct {
	index    library.GeoIndex
	resolver geo.Geocoder
}

func NewGeocoder(idx library.GeoIndex) *Geocoder {
	return &Geocoder{
		index:    idx,
		resolver: openstreetmap.Geocoder(),
	}
}

func (g *Geocoder) RegisterTasks(repo *tasks.TaskRepository) {
	repo.Register("geoResolve", func() tasks.Task {
		return NewGeoLookupTask(g)
	})
}

func (g *Geocoder) ReverseGeocode(lat, lon float64) (*gps.Address, error) {
	address, err := g.resolver.ReverseGeocode(lat, lon)
	if err != nil {
		return nil, err
	}
	if address == nil {
		return nil, UnknownLocation
	}
	gpsAddess := toAddress(*address)
	return &gpsAddess, nil
}

func (g *Geocoder) LookupPhotoOnAdd(ctx context.Context, p *library.Photo) (tasks.Task, bool) {
	if p.Location == nil || !p.Location.IsValid() {
		return nil, false
	}
	return NewGeoLookupTaskWith(g, p.ID, *p.Location), true
}

func (g *Geocoder) ResolveAndStoreLocation(ctx context.Context, p library.PhotoID, coords gps.Coordinates) error {
	logger, ctx := logging.FromWithNameAndFields(ctx, "geocoder", zap.String("photo", string(p)))
	logger.Info("Reverse geocoding", zap.Stringer("location", coords))
	address, err := g.ReverseGeocode(coords.Lat, coords.Long)
	if err != nil {
		return err
	}
	if address == nil {
		return nil
	}
	logger.Info("Geo decoded: ",
		zap.Stringer("country", address.Country.ID),
		zap.String("city", address.City))
	return g.index.Update(ctx, p, address)
}
