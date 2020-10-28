package geocoding

import (
	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/codingsince1985/geo-golang"
)

func toAddress(address geo.Address) (out gps.Address) {
	return gps.AsAddress(address.Country, address.CountryCode, address.City, address.Postcode)
}
