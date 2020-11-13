package gps

import (
	"errors"
	"fmt"
)

var (
	InvalidGPSCoordinates = errors.New("Invalid GPS coordinates")
)

type Coordinates struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

const (
	LatMin = float64(-90.)
	LatMax = float64(90.)
	LonMin = float64(-180.)
	LonMax = float64(180.)
)

var (
	WorldBounds = Rect{LonMin, LatMin, LonMax, LatMax}
)

func NewCoordinates(lat, lon float64) (*Coordinates, error) {
	if lat < LatMin || lat > LatMax || lon < LonMin || lon > LonMax {
		return nil, InvalidGPSCoordinates
	}
	return &Coordinates{Lat: lat, Long: lon}, nil
}

func MustNewCoordinates(lat, lon float64) *Coordinates {
	if lat < LatMin || lat > LatMax || lon < LonMin || lon > LonMax {
		panic(fmt.Sprintf("Invalid GPS coordinates [%f;%f]", lat, lon))
	}
	return &Coordinates{lat, lon}
}

func (c Coordinates) IsValid() bool {
	return c.Lat >= LatMin && c.Lat <= LatMax && c.Long >= LonMin && c.Long <= LonMax
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
