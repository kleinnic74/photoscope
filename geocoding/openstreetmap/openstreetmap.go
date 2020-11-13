package openstreetmap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/geocoding"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

var IllegalBoundingBox = errors.New("Not a valid bounding box")

const (
	baseURL   = "https://nominatim.openstreetmap.org/"
	userAgent = "GOPhotos/0.1"
)

type resolver struct {
	lang   string
	client http.Client
}

type boundingbox gps.Rect

func (b *boundingbox) Rect() *gps.Rect {
	if b == nil {
		return nil
	} else {
		var r gps.Rect
		r = gps.Rect(*b)
		return &r
	}
}

func (b *boundingbox) UnmarshalJSON(data []byte) error {
	var points []string
	if err := json.Unmarshal(data, &points); err != nil {
		return nil
	}
	if len(points) == 0 {
		return nil
	}
	if len(points) != 4 {
		return IllegalBoundingBox
	}
	var err error
	b[0], err = strconv.ParseFloat(points[2], 64)
	b[1], err = strconv.ParseFloat(points[0], 64)
	b[2], err = strconv.ParseFloat(points[3], 64)
	b[3], err = strconv.ParseFloat(points[1], 64)
	return err
}

type address struct {
	City       string `json:"city"`
	Zip        string `json:"postcode"`
	Country    string `json:"state"`
	CountryISO string `json:"country_code"`
}

type latlon float64

func (p *latlon) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return err
	}
	*p = latlon(f)
	return nil
}

type location struct {
	ID          int          `json:"osm_id"`
	OSMType     string       `json:"osm_type"`
	Type        string       `json:"type"`
	Lat         latlon       `json:"lat"`
	Long        latlon       `json:"lon"`
	BoundingBox *boundingbox `json:"boundingbox"`
	DisplayName string       `json:"display_name"`
	Address     address      `json:"address"`
}

func (l location) Pos() gps.Point {
	return gps.Point{float64(l.Long), float64(l.Lat)}
}

func NewResolver(lang ...string) geocoding.Resolver {
	return &resolver{
		lang: strings.Join(lang, ","),
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (osm *resolver) ReverseGeocode(ctx context.Context, lat, long float64) (*gps.Address, bool, error) {
	logger, ctx := logging.SubFrom(ctx, "openstreetmap")
	url := fmt.Sprintf("%s/reverse?format=json&lat=%f&lon=%f&addressdetails=1&zoom=14", baseURL, lat, long)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", osm.lang)
	res, err := osm.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
	}
	logger.Debug("reverseGeocode response", zap.String("response", string(data)))
	var location location
	if err := json.Unmarshal(data, &location); err != nil {
		return nil, false, err
	}
	return &gps.Address{
		AddressFields: gps.AddressFields{
			Country: gps.Country{
				Country: location.Address.Country,
				ID:      gps.CountryID(location.Address.CountryISO),
			},
			City: location.Address.City,
			Zip:  location.Address.Zip,
		},
		BoundingBox: location.BoundingBox.Rect(),
	}, true, nil
}
