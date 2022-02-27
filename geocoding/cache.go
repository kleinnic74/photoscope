package geocoding

import (
	"context"
	"sync"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type Stats struct {
	Hits         int `json:"hits"`
	Misses       int `json:"misses"`
	MultiMatches int `json:"multimatches"`
	Total        int `json:"total"`
}

type internalStats struct {
	Stats
	hits         prometheus.Counter
	misses       prometheus.Counter
	multiMatches prometheus.Counter
	total        prometheus.Counter
}

func (s *internalStats) hit() {
	s.total.Inc()
	s.Stats.Total++
	s.hits.Inc()
	s.Stats.Hits++
}

func (s *internalStats) miss() {
	s.total.Inc()
	s.Stats.Total++
	s.misses.Inc()
	s.Stats.Misses++
}

func (s *internalStats) multiMatch() {
	s.total.Inc()
	s.Stats.Total++
	s.multiMatches.Inc()
	s.Stats.MultiMatches++
	s.misses.Inc()
	s.Stats.Misses++
}

type Cache struct {
	stats    *internalStats
	delegate Resolver

	qt   *quadtree
	lock sync.RWMutex
}

func NewGeoCache(r Resolver) *Cache {
	stats := &internalStats{
		hits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "geocoding_cache_hits",
			Help: "Number of addresses successfully resolved through the cache",
		}),
		misses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "geocoding_cache_misses",
			Help: "Number of reverse geocoding requests not found in the cache",
		}),
		multiMatches: promauto.NewCounter(prometheus.CounterOpts{
			Name: "geocoding_cache_multimatches",
			Help: "Number of multiple matches found in cache for a given coordinate, also count as miss",
		}),
		total: promauto.NewCounter(prometheus.CounterOpts{
			Name: "geocoding_cache_total",
			Help: "Total number of requests to the geocoding cache",
		}),
	}
	return &Cache{stats: stats, delegate: r, qt: NewQuadTree(gps.WorldBounds), lock: sync.RWMutex{}}
}

func (c *Cache) DumpStats() Stats {
	return c.stats.Stats
}

func (c *Cache) ReverseGeocode(ctx context.Context, lat, lon float64) (*gps.Address, bool, error) {
	log, ctx := logging.FromWithNameAndFields(ctx, "geocache", zap.Stringer("pos", gps.PointFromLatLon(lat, lon)))
	places := c.findPlace(lat, lon)
	switch len(places) {
	case 1:
		c.stats.hit()
		return places[0].(*gps.Address), true, nil
	case 0:
		c.stats.miss()
		log.Debug("No place found in cache")
	default:
		c.stats.multiMatch()
		log.Debug("Multiple places found in cache")
	}
	place, found, err := c.delegate.ReverseGeocode(ctx, lat, lon)
	if found && place.BoundingBox != nil {
		c.add(*place.BoundingBox, place)
	} else if found {
		log.Info("Place has no bounding box", zap.Stringer("place", place.ID))
	} else {
		log.Info("Place not found", zap.Float64("lat", lat), zap.Float64("long", lon))
	}
	return place, found, err
}

func (c *Cache) findPlace(lat float64, lon float64) []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.qt.Find(gps.PointFromLatLon(lat, lon))
}

func (c *Cache) Add(place gps.Address) bool {
	if !place.HasValidBoundingBox() {
		return false
	}
	c.add(*place.BoundingBox, &place)
	return true
}

func (c *Cache) add(r gps.Rect, place *gps.Address) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.qt.InsertRect(r, place)
}

func (c *Cache) Visit(v Visitor) {
	c.qt.Visit(v)
}
