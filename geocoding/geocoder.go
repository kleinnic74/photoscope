package geocoding

import (
	"context"
	"errors"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/tasks"
	"go.uber.org/zap"
)

var UnknownLocation = errors.New("Unknown GPS location")

type Resolver interface {
	ReverseGeocode(ctx context.Context, lat, long float64) (*gps.Address, bool, error)
}

type Geocoder struct {
	index    library.GeoIndex
	resolver Resolver
	cache    *cache
}

func NewGeocoder(idx library.GeoIndex, resolver Resolver) *Geocoder {
	c := NewGeoCache(resolver)
	return &Geocoder{
		index:    idx,
		resolver: c,
		cache:    c,
	}
}

func (g *Geocoder) RegisterTasks(repo *tasks.TaskRepository) {
	repo.Register("geoResolve", func() tasks.Task {
		return NewGeoLookupTask(g)
	})
	repo.RegisterWithProperties("populateCache", func() tasks.Task {
		return newLoadKnownPlaces(g.index, g.cache)
	}, tasks.TaskProperties{
		RunOnStart:   true,
		UserRunnable: false,
	})
}

func (g *Geocoder) ReverseGeocode(ctx context.Context, lat, lon float64) (*gps.Address, error) {
	address, found, err := g.resolver.ReverseGeocode(ctx, lat, lon)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, UnknownLocation
	}
	return address, nil
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
	address, err := g.ReverseGeocode(ctx, coords.Lat, coords.Long)
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
