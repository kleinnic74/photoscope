package geocoding

import (
	"context"
	"sync"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

type cache struct {
	delegate Resolver

	qt   *quadtree
	lock sync.RWMutex

	Hits   int
	Misses int
}

func NewGeoCache(r Resolver) *cache {
	return &cache{delegate: r, qt: NewQuadTree(gps.WorldBounds), lock: sync.RWMutex{}}
}

func (c *cache) ReverseGeocode(ctx context.Context, lat, lon float64) (*gps.Address, bool, error) {
	log, ctx := logging.FromWithNameAndFields(ctx, "geocache")
	places := c.findPlace(lat, lon)
	if len(places) == 1 {
		c.Hits++
		return places[0].(*gps.Address), true, nil
	}
	c.Misses++
	log.Debug("No place found in cache", zap.Stringer("pos", gps.PointFromLatLon(lat, lon)))
	place, found, err := c.delegate.ReverseGeocode(ctx, lat, lon)
	if found && place.BoundingBox != nil {
		c.add(*place.BoundingBox, place)
	} else {
		log.Info("Place has no bounding box", zap.Stringer("place", place.ID))
	}
	return place, found, err
}

func (c *cache) findPlace(lat float64, lon float64) []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.qt.Find(gps.PointFromLatLon(lat, lon))
}

func (c *cache) Add(place gps.Address) {
	if !place.HasValidBoundingBox() {
		return
	}
	c.add(*place.BoundingBox, &place)
}

func (c *cache) add(r gps.Rect, place *gps.Address) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.qt.InsertRect(r, place)
}

func (c *cache) Visit(v Visitor) {
	c.qt.Visit(v)
}
