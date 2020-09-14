package geocoding

import (
	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/codingsince1985/geo-golang"
)

func toAddress(address geo.Address) (out gps.Address) {
	return gps.Address{
		Country:     address.Country,
		CountryCode: address.CountryCode,
		City:        address.City,
		Zip:         address.Postcode,
		County:      address.County,
	}
}
