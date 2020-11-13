package openstreetmap

import (
	"context"
	"encoding/json"
	"testing"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/stretchr/testify/assert"
)

const response = `{"place_id":47300855,"licence":"Data © OpenStreetMap contributors, ODbL 1.0. https://osm.org/copyright","osm_type":"node","osm_id":3617499243,"lat":"48.2118494","lon":"16.3651666","display_name":"Schottenviertel, KG Innere Stadt, Innere Stadt, Wien, 1010, Austria","address":{"neighbourhood":"Schottenviertel","suburb":"KG Innere Stadt","city_district":"Innere Stadt","city":"Wien","postcode":"1010","country":"Austria","country_code":"at"},"boundingbox":["48.2018494","48.2218494","16.3551666","16.3751666"]}`

func TestUnmarshalBoundingbox(t *testing.T) {
	// Format is: [lat0, lat1, lon0, lon1]
	boxjson := `["12.34", "23.45", "34.56", "45.67"]`
	var bbox boundingbox
	err := json.Unmarshal([]byte(boxjson), &bbox)
	if err != nil {
		t.Fatalf("Error unmarshaling JSON: %s", err)
	}
	assert.Equal(t, 34.56, bbox[0])
	assert.Equal(t, 12.34, bbox[1])
	assert.Equal(t, 45.67, bbox[2])
	assert.Equal(t, 23.45, bbox[3])
}

func TestUnmarshalResponse(t *testing.T) {
	var l location
	err := json.Unmarshal([]byte(response), &l)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %s", err)
	}
	assert.Equal(t, latlon(48.2118494), l.Lat, "Bad value for lattitude")
	assert.Equal(t, latlon(16.3651666), l.Long, "Bad value for longitude")
	assert.NotNil(t, l.BoundingBox)
	pos := l.Pos()
	assert.True(t, pos.In(gps.Rect(*l.BoundingBox)))
	assert.Equal(t, gps.RectFrom(16.3551666, 48.2018494, 16.3751666, 48.2218494), gps.Rect(*l.BoundingBox))
}

func TestReverseGeocode(t *testing.T) {
	osm := NewResolver("de,en")
	address, found, err := osm.ReverseGeocode(context.Background(), 52.5487429714954, -1.81602098644987)
	if err != nil {
		t.Fatalf("Error while reverse geocoding: %s", err)
	}
	if !found {
		t.Errorf("Bad result, found was %t, expected %t", found, true)
	}
	t.Logf("Resolved: %v", address)
}
