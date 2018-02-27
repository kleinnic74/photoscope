package domain

import "fmt"

type Coordinates struct {
	lat  float64
	long float64
}

func (c Coordinates) String() string {
	return fmt.Sprintf("[%f;%f]", c.lat, c.long)
}

func (c Coordinates) DistanceTo(other *Coordinates) float64 {
	return 0
}
