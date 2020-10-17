package boltstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/library/index"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const GeoIndexVersion = index.Version(2)

type boltGeoIndex struct {
	db *bolt.DB
}

var (
	// placeOfPhotos tracks the location of each photo, indexed by PhotoID
	placeOfPhotos = []byte("photoplaces")
	// photosByPlace tracks the photos at a given place, indexed by place key
	photosByPlace = []byte("photosByPlace")
	// allCountriesBucket contains all known countries, indexed by country code
	allCountriesBucket = []byte("allcountries")
	// placesByCountryBucket stores all places in a given country, indexed by country code
	placesByCountryBucket = []byte("placesByCountry")
)

const (
	unknownPlacesKey = "_unknown"
)

func NewBoltGeoIndex(db *bolt.DB) (library.GeoIndex, error) {
	if err := createBucket(db, placeOfPhotos); err != nil {
		return nil, err
	}
	if err := createBucket(db, allCountriesBucket); err != nil {
		return nil, err
	}
	if err := createBucket(db, placesByCountryBucket); err != nil {
		return nil, err
	}
	if err := createBucket(db, photosByPlace); err != nil {
		return nil, err
	}
	return &boltGeoIndex{
		db: db,
	}, nil
}

func (index *boltGeoIndex) Has(ctx context.Context, id library.PhotoID) (exists bool) {
	logger, ctx := logging.FromWithNameAndFields(ctx, "geoStore")
	err := index.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(placeOfPhotos)
		exists = b.Get([]byte(id)) != nil
		return nil
	})
	if err != nil {
		logger.Warn("Bolt error", zap.String("photo", string(id)), zap.Error(err))
	}
	return
}

func (index *boltGeoIndex) Get(ctx context.Context, id library.PhotoID) (address *gps.Address, found bool, err error) {
	logger, ctx := logging.FromWithNameAndFields(ctx, "geoStore")
	if err := index.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(placeOfPhotos)
		data := b.Get([]byte(id))
		found = data != nil
		if !found {
			return nil
		}
		address = new(gps.Address)
		return json.Unmarshal(data, address)
	}); err != nil {
		logger.Warn("Bolt error", zap.String("photo", string(id)), zap.Error(err))
		return nil, false, err
	}
	return
}

func (index *boltGeoIndex) Update(ctx context.Context, id library.PhotoID, address *gps.Address) error {
	if address == nil {
		return nil
	}
	encodedAddress, err := json.Marshal(address)
	if err != nil {
		return err
	}
	return index.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(placeOfPhotos)
		if err := b.Put([]byte(id), encodedAddress); err != nil {
			return err
		}
		allCountries := tx.Bucket(allCountriesBucket)
		// Save country object
		encodedCountry, err := json.Marshal(&address.Country)
		if err != nil {
			return err
		}
		allCountries.Put([]byte(address.Country.Code), encodedCountry)

		placesByCountry := tx.Bucket(placesByCountryBucket)
		placesInCountryBucketName := []byte(address.Country.Code)

		placesInCountry, err := placesByCountry.CreateBucketIfNotExists(placesInCountryBucketName)
		if err != nil {
			return err
		}
		fullKey, inCountryKey := placeKeys(*address)
		if err := placesInCountry.Put(inCountryKey, encodedAddress); err != nil {
			return err
		}

		photosByPlace := tx.Bucket(photosByPlace)
		photosAtPlace, err := photosByPlace.CreateBucketIfNotExists(fullKey)
		if err != nil {
			return err
		}
		return photosAtPlace.Put([]byte(id), []byte(id))
	})
}

func fullPlaceKey(country string, zip string) string {
	country = strings.ToUpper(country)
	zip = strings.ToLower(zip)
	return fmt.Sprintf("%s/%s", country, zip)
}

func placeKeys(place gps.Address) ([]byte, []byte) {
	var placeKey string
	if place.Zip != "" {
		placeKey = place.Zip
	} else {
		placeKey = unknownPlacesKey
	}
	fullKey := fullPlaceKey(place.Code, placeKey)
	return []byte(fullKey), []byte(placeKey)
}

func (index *boltGeoIndex) Locations(ctx context.Context) (*library.Locations, error) {
	var locations library.Locations
	err := index.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(allCountriesBucket)
		return b.ForEach(func(k, v []byte) error {
			var country gps.Country
			if err := json.Unmarshal(v, &country); err != nil {
				return err
			}
			var countryAndPlaces library.CountryAndPlaces
			countryAndPlaces.Country = country

			placesByCountry := tx.Bucket(placesByCountryBucket)
			if places := placesByCountry.Bucket([]byte(country.Code)); places != nil {
				c := places.Cursor()
				for k, v := c.First(); k != nil; k, v = c.Next() {
					var address gps.Address
					if err := json.Unmarshal(v, &address); err != nil {
						return err
					}
					countryAndPlaces.Places = append(countryAndPlaces.Places, &address.Place)
				}
			}
			locations.Countries = append(locations.Countries, &countryAndPlaces)
			return nil
		})
	})
	return &locations, err
}

func (index *boltGeoIndex) FindByPlacePaged(ctx context.Context, country string, zip string, startAt int, maxCount int) (photos []library.PhotoID, hasMore bool, err error) {
	err = index.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosByPlace)
		key := fullPlaceKey(country, zip)
		sub := b.Bucket([]byte(key))
		if sub == nil {
			return nil
		}
		c := sub.Cursor()
		var index int
		var count int
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if index < startAt {
				index++
				continue
			}
			if count >= maxCount {
				hasMore = true
				return nil
			}
			photos = append(photos, library.PhotoID(k))
			count++
		}
		return nil
	})
	return
}

func (index *boltGeoIndex) FindByCountryPaged(ctx context.Context, country string, startAt int, maxCount int) (photos []library.PhotoID, hasMore bool, err error) {
	err = index.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(photosByPlace)
		keyPrefix := []byte(fmt.Sprintf("%s/", country))
		buckets := b.Cursor()
		var index int
		var count int
		for subK, _ := buckets.Seek(keyPrefix); subK != nil && bytes.HasPrefix(subK, keyPrefix); subK, _ = buckets.Next() {
			sub := b.Bucket(subK)
			if sub == nil {
				continue
			}
			c := sub.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				if index < startAt {
					index++
					continue
				}
				if count >= maxCount {
					hasMore = true
					return nil
				}
				photos = append(photos, library.PhotoID(k))
				count++
			}
		}
		return nil
	})
	return
}
