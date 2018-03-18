package gps

import (
	"encoding/json"
	"testing"
)

func TestCoordinatesToISO6709(t *testing.T) {
	var data = []struct {
		lat  float64
		long float64
		iso  string
	}{
		{lat: 45.3, long: 2.443, iso: "+45.300000+002.443000/"},
		{lat: 45.3, long: -43.2344, iso: "+45.300000-043.234400/"},
	}
	for _, tt := range data {
		c := NewCoordinates(tt.lat, tt.long)
		iso := c.ISO6709()
		if iso != tt.iso {
			t.Errorf("Bad ISO6709 value, expected %s, got %s", tt.iso, iso)
		}
	}
}

func BenchmarkMNarshalJSON(b *testing.B) {
	coords := Coordinates{
		lat:  12.3456,
		long: 23.2344,
	}
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(coords)
		if err != nil {
			b.Error(err)
		}
	}
}
