package library

import (
	"fmt"
	"math/rand"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/gps"
)

func RandomPhoto() *Photo {
	id := rand.Uint32()
	dateTaken, _ := time.Parse(time.RFC3339, "2018-02-23T13:43:12Z")
	f := domain.MustFormatForExt("jpg")
	return &Photo{
		ID:          PhotoID(fmt.Sprintf("%8d", id)),
		Path:        "2018/02/23",
		Format:      f,
		DateTaken:   dateTaken,
		Orientation: 1,
		Location:    gps.NewCoordinates(47.123445, 45.12313),
	}
}
