package gps

import (
	"fmt"
)

var (
	Unknown *Coordinates
)

type Coordinates struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

func NewCoordinates(lat, long float64) *Coordinates {
	return &Coordinates{Lat: lat, Long: long}
}

func (c Coordinates) String() string {
	return fmt.Sprintf("[%f;%f]", c.Lat, c.Long)
}

func (c Coordinates) DistanceTo(other *Coordinates) float64 {
	return 0
}

func (c *Coordinates) ISO6709() string {
	return fmt.Sprintf("%+010.6f%+011.6f/", c.Lat, c.Long)
}

func init() {
	Unknown = &Coordinates{0, 0}
}
