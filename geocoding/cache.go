package geocoding

import (
	"context"
	"log"

	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type cache struct {
	delegate Resolver

	qt *quadtree

	Hits   int
	Misses int
}

func NewGeoCache(r Resolver) *cache {
	return &cache{delegate: r, qt: NewQuadTree(gps.WorldBounds)}
}

func (c *cache) ReverseGeocode(ctx context.Context, lat, lon float64) (*gps.Address, bool, error) {
	places := c.qt.Find(gps.PointFromLatLon(lat, lon))
	if len(places) == 1 {
		c.Hits++
		return places[0].(*gps.Address), true, nil
	}
	c.Misses++
	log.Printf("No place found in cache (%d items in cache)", c.qt.count)
	place, found, err := c.delegate.ReverseGeocode(ctx, lat, lon)
	if found && place.BoundingBox != nil {
		c.qt.InsertRect(*place.BoundingBox, place)
	} else {
		log.Printf("Place has no bounding box: %v", place.BoundingBox)
	}
	return place, found, err
}

func (c *cache) Visit(v Visitor) {
	c.qt.Visit(v)
}
