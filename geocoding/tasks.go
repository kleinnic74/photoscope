package geocoding

import (
	"context"
	"fmt"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/tasks"
	"go.uber.org/zap"

	"github.com/codingsince1985/geo-golang"
	"github.com/codingsince1985/geo-golang/openstreetmap"
)

var resolver geo.Geocoder

func RegisterTasks(repo *tasks.TaskRepository, index library.GeoIndex) {
	resolver = openstreetmap.Geocoder()
	repo.RegisterWithProperties(
		"scanForGeoUnresolvedPhotos",
		func() tasks.Task {
			return NewLocationScannerTask(index)
		},
		tasks.TaskProperties{
			RunOnStart: true,
		},
	)
	repo.Register("geoResolve", func() tasks.Task {
		return NewGeoLookupTask(index)
	})
}

type locationScannerTask struct {
	geoindex library.GeoIndex
}

type geoLookupTask struct {
	PhotoID  string          `json:"photoID"`
	Coords   gps.Coordinates `json:"gps"`
	geoindex library.GeoIndex
}

func NewLocationScannerTask(index library.GeoIndex) tasks.Task {
	return locationScannerTask{geoindex: index}
}

func (t locationScannerTask) Describe() string {
	return "Looking for photos with unresolved location data"
}

func (t locationScannerTask) Execute(ctx context.Context, executor tasks.TaskExecutor, lib library.PhotoLibrary) error {
	logger, ctx := logging.SubFrom(ctx, "locationScanner")
	photos, err := lib.FindAll(ctx)
	if err != nil {
		return err
	}
	var count int
	for _, p := range photos {
		if p.Location() == nil {
			continue
		}
		if t.geoindex.Has(ctx, p.ID()) {
			continue
		}
		logger.Info("LocationUpgradeNeeded", zap.String("photo", p.ID()))
		executor.Submit(ctx, NewGeoLookupTaskWith(t.geoindex, p.ID(), *p.Location()))
		count++
	}
	logger.Info("Location scan done", zap.Int("needLookup", count))
	return nil
}

func NewGeoLookupTask(index library.GeoIndex) tasks.Task {
	return geoLookupTask{geoindex: index}
}

func NewGeoLookupTaskWith(index library.GeoIndex, id string, coords gps.Coordinates) tasks.Task {
	return geoLookupTask{
		PhotoID:  id,
		Coords:   coords,
		geoindex: index,
	}
}

func (t geoLookupTask) Describe() string {
	return fmt.Sprintf("Looking up location of photo %s", t.PhotoID)
}

func (t geoLookupTask) Execute(ctx context.Context, executor tasks.TaskExecutor, lib library.PhotoLibrary) error {
	logger, ctx := logging.FromWithNameAndFields(ctx, "geoLookupTask", zap.String("photo", t.PhotoID))
	logger.Info("Reverse geocoding", zap.Stringer("location", t.Coords))
	address, err := resolver.ReverseGeocode(t.Coords.Lat, t.Coords.Long)
	if err != nil {
		return err
	}
	if address == nil {
		return nil
	}
	a := toAddress(*address)
	logger.Info("Geo decoded: ",
		zap.String("country", a.Country),
		zap.String("city", a.City),
		zap.String("county", a.County))
	return t.geoindex.Update(ctx, t.PhotoID, &a)
}
