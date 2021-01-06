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
	coords, _ := gps.NewCoordinates(47.123445, 45.12313)
	photoID := PhotoID(fmt.Sprintf("%8d", id))
	return &Photo{
		ExtendedPhotoID: ExtendedPhotoID{
			ID:     photoID,
			SortID: orderedIDOf(dateTaken, photoID),
		},
		Path:        "2018/02/23",
		Format:      f,
		DateTaken:   dateTaken,
		Orientation: 1,
		Location:    coords,
	}
}
