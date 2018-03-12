package gps

import (
	"encoding/json"
	"fmt"
)

var (
	Unknown *Coordinates
)

type Coordinates struct {
	lat  float64
	long float64
}

func (gps *Coordinates) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Lat  float64 `json:"lat"`
		Long float64 `json:"long"`
	}{
		Lat:  gps.lat,
		Long: gps.long,
	})
}

func (gps *Coordinates) UnmarshalJSON(buf []byte) error {
	var c struct {
		Lat  float64 `json:"lat"`
		Long float64 `json:"long"`
	}
	if err := json.Unmarshal(buf, &c); err != nil {
		return err
	}
	gps.lat = c.Lat
	gps.long = c.Long
	return nil
}

func NewCoordinates(lat, long float64) Coordinates {
	return Coordinates{lat: lat, long: long}
}

func (c Coordinates) String() string {
	return fmt.Sprintf("[%f;%f]", c.lat, c.long)
}

func (c Coordinates) DistanceTo(other *Coordinates) float64 {
	return 0
}

func (c *Coordinates) ISO6709() string {
	return fmt.Sprintf("%+010.6f%+011.6f/", c.lat, c.long)
}

func init() {
	Unknown = &Coordinates{0, 0}
}
