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
	latMin  = float64(-90.)
	latMax  = float64(90.)
	longMin = float64(-180.)
	longMax = float64(180.)
)

func NewCoordinates(lat, lon float64) (*Coordinates, error) {
	if lat < latMin || lat > latMax || lon < longMin || lon > longMax {
		return nil, InvalidGPSCoordinates
	}
	return &Coordinates{Lat: lat, Long: lon}, nil
}

func MustNewCoordinates(lat, lon float64) *Coordinates {
	if lat < latMin || lat > latMax || lon < longMin || lon > longMax {
		panic(fmt.Sprintf("Invalid GPS coordinates [%f;%f]", lat, lon))
	}
	return &Coordinates{lat, lon}
}

func (c Coordinates) IsValid() bool {
	return c.Lat >= latMin && c.Lat <= latMax && c.Long >= longMin && c.Long <= longMax
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
