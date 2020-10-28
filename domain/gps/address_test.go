package gps

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var addressData = []struct {
	JSON   string
	STRUCT Address
	ID     PlaceID
}{
	{JSON: `{"country":"France","ciso":"fr","city":"Nice","zip":"06000"}`,
		STRUCT: Address{
			AddressFields: AddressFields{
				Country: Country{Country: "France", ID: CountryIDFromString("FR")},
				City:    "Nice",
				Zip:     "06000",
			},
		},
		ID: PlaceID("fr/06000/nice"),
	},
}

func TestUnmarshalAddress(t *testing.T) {
	for i, d := range addressData {
		var a Address
		if err := json.Unmarshal([]byte(d.JSON), &a); err != nil {
			t.Fatalf("#%d: error while unmarshalling JSON: %s", i, err)
		}
		assert.Equal(t, d.STRUCT.AddressFields, a.AddressFields)
		assert.Equal(t, d.ID, a.ID)
	}
}

func TestMarshalJSONAddress(t *testing.T) {
	for i, d := range addressData {
		bin, err := json.Marshal(&d.STRUCT)
		if err != nil {
			t.Fatalf("#%d: error while marshalling to JSON: %s", i, err)
		}
		assert.Equal(t, d.JSON, string(bin))
	}
}
