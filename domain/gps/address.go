package gps

import (
	"encoding/json"
	"strings"
)

type PlaceID string

func (id PlaceID) String() string {
	return string(id)
}

type Country struct {
	Country string    `json:"country,omitempty"`
	ID      CountryID `json:"ciso,omitempty"`
}

type CountryID string

func (id CountryID) String() string {
	return string(id)
}

func CountryIDFromString(id string) CountryID {
	return CountryID(strings.ToLower(id))
}

type AddressFields struct {
	Country
	City string `json:"city,omitempty"`
	Zip  string `json:"zip,omitempty"`
}

// Address is the address view of a geographical location
type Address struct {
	AddressFields
	ID          PlaceID `json:"id"`
	BoundingBox *Rect   `json:"boundingbox"`
}

func AsAddress(country, iso, city, zip string) Address {
	cid := CountryIDFromString(iso)
	return Address{
		AddressFields: AddressFields{
			Country: Country{
				Country: country,
				ID:      cid,
			},
			City: city,
			Zip:  zip,
		},
		ID: asPlaceID(cid, city, zip),
	}
}

func (a *Address) UnmarshalJSON(data []byte) (err error) {
	var fields AddressFields
	if err = json.Unmarshal(data, &fields); err != nil {
		return
	}
	a.ID = asPlaceID(fields.Country.ID, fields.City, fields.Zip)
	a.AddressFields = fields
	return
}

func asPlaceID(countryCode CountryID, city, zip string) PlaceID {
	return PlaceID(strings.Join([]string{strings.ToLower(string(countryCode)), strings.ToLower(zip), strings.ToLower(city)}, "_"))
}
