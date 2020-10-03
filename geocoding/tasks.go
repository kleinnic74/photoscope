package geocoding

import (
	"context"
	"fmt"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/tasks"
	"go.uber.org/zap"

	"github.com/codingsince1985/geo-golang/openstreetmap"
)

var resolver = openstreetmap.Geocoder()

func RegisterTasks(repo *tasks.TaskRepository, index library.GeoIndex) {
	repo.Register("geoResolve", func() tasks.Task {
		return NewGeoLookupTask(index)
	})
}

type geoLookupTask struct {
	PhotoID  library.PhotoID `json:"photoID"`
	Coords   gps.Coordinates `json:"gps"`
	geoindex library.GeoIndex
}

func NewGeoLookupTask(index library.GeoIndex) tasks.Task {
	return geoLookupTask{geoindex: index}
}

func NewGeoLookupTaskWith(index library.GeoIndex, id library.PhotoID, coords gps.Coordinates) tasks.Task {
	return geoLookupTask{
		PhotoID:  id,
		Coords:   coords,
		geoindex: index,
	}
}

func LookupPhotoOnAdd(index library.GeoIndex) tasks.DeferredNewPhotoCallback {
	return func(ctx context.Context, p *library.Photo) (tasks.Task, bool) {
		if p.Location == nil {
			return nil, false
		}
		return NewGeoLookupTaskWith(index, p.ID, *p.Location), true
	}
}

func (t geoLookupTask) Describe() string {
	return fmt.Sprintf("Looking up location of photo %s", t.PhotoID)
}

func (t geoLookupTask) Execute(ctx context.Context, executor tasks.TaskExecutor, lib library.PhotoLibrary) error {
	logger, ctx := logging.FromWithNameAndFields(ctx, "geoLookupTask", zap.String("photo", string(t.PhotoID)))
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
