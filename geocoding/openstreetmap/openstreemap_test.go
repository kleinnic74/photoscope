package openstreetmap

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/stretchr/testify/assert"
)

var data = []struct {
	lat      float64
	lon      float64
	response string
	out      gps.AddressFields
}{
	{
		lat:      52.5487429714954,
		lon:      -1.81602098644987,
		response: `{"place_id":47300855,"licence":"Data © OpenStreetMap contributors, ODbL 1.0. https://osm.org/copyright","osm_type":"node","osm_id":3617499243,"lat":"48.2118494","lon":"16.3651666","display_name":"Schottenviertel, KG Innere Stadt, Innere Stadt, Wien, 1010, Austria","address":{"neighbourhood":"Schottenviertel","suburb":"KG Innere Stadt","city_district":"Innere Stadt","city":"Wien","postcode":"1010","country":"Austria","country_code":"at"},"boundingbox":["48.2018494","48.2218494","16.3551666","16.3751666"]}`,
		out:      gps.AddressFields{Country: gps.Country{ID: "at", Country: "Austria"}, City: "Wien", Zip: "1010"},
	},
	{
		lat:      48.081489,
		lon:      15.592614,
		response: "{\"place_id\":94267472,\"licence\":\"Data © OpenStreetMap contributors, ODbL 1.0. https://osm.org/copyright\",\"osm_type\":\"way\",\"osm_id\":29301551,\"lat\":\"48.08147239671421\",\"lon\":\"15.592610324427119\",\"display_name\":\"Stelzhamergasse, Göblasbruck, Gemeinde Wilhelmsburg, Bezirk St. Pölten, Niederösterreich, 3150, Austria\",\"address\":{\"road\":\"Stelzhamergasse\",\"residential\":\"Göblasbruck\",\"suburb\":\"Göblasbruck\",\"town\":\"Gemeinde Wilhelmsburg\",\"county\":\"Bezirk St. Pölten\",\"state\":\"Niederösterreich\",\"postcode\":\"3150\",\"country\":\"Austria\",\"country_code\":\"at\"},\"boundingbox\":[\"48.0812307\",\"48.081669\",\"15.5916536\",\"15.5938144\"]}",
		out:      gps.AddressFields{Country: gps.Country{ID: "at", Country: "Austria"}, City: "Gemeinde Wilhelmsburg", Zip: "3150"},
	},
}

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
	err := json.Unmarshal([]byte(data[0].response), &l)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %s", err)
	}
	assert.Equal(t, latlon(48.2118494), l.Lat, "Bad value for lattitude")
	assert.Equal(t, latlon(16.3651666), l.Long, "Bad value for longitude")
	assert.Equal(t, "1010", l.Address.Zip)
	assert.Equal(t, "Austria", l.Address.Country)
	assert.NotNil(t, l.BoundingBox)
	pos := l.Pos()
	assert.True(t, pos.In(gps.Rect(*l.BoundingBox)))
	assert.Equal(t, gps.RectFrom(16.3551666, 48.2018494, 16.3751666, 48.2218494), gps.Rect(*l.BoundingBox))
}

func TestReverseGeocode(t *testing.T) {
	for _, d := range data {
		testClient := newTestClient(func(r *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBufferString(d.response)),
			}
		})
		osm := NewResolverWithClient(testClient, "de", "en")
		address, found, err := osm.ReverseGeocode(context.Background(), d.lat, d.lon)
		if err != nil {
			t.Fatalf("Error while reverse geocoding: %s", err)
		}
		if !found {
			t.Errorf("Bad result, found was %t, expected %t", found, true)
		}
		t.Logf("Resolved: %v", address)
		assert.NotEmpty(t, address.ID)
		assert.Equal(t, d.out.Zip, address.Zip)
		assert.Equal(t, d.out.City, address.City)
		assert.Equal(t, d.out.Country, address.Country)
	}
}

type RoundTripperFunc func(*http.Request) *http.Response

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(roundTripFunc RoundTripperFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc,
	}
}
