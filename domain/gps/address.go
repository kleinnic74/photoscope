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
	BoundingBox *Rect   `json:"boundingbox,omitempty"`
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
	type aliasAddress *Address
	alias := aliasAddress(a)
	if err = json.Unmarshal(data, alias); err != nil {
		return
	}
	a.ID = asPlaceID(alias.AddressFields.Country.ID, alias.AddressFields.City, alias.AddressFields.Zip)
	return
}

func (a *Address) HasValidBoundingBox() bool {
	return a.BoundingBox != nil && a.BoundingBox.W() > 0 && a.BoundingBox.H() > 0
}

func asPlaceID(countryCode CountryID, city, zip string) PlaceID {
	return PlaceID(strings.Join([]string{strings.ToLower(string(countryCode)), strings.ToLower(zip), strings.ToLower(city)}, "_"))
}
