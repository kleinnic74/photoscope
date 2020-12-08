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
	{JSON: `{"country":"France","ciso":"fr","city":"Nice","zip":"06000","id":"fr_06000_nice"}`,
		STRUCT: Address{
			AddressFields: AddressFields{
				Country: Country{Country: "France", ID: CountryIDFromString("FR")},
				City:    "Nice",
				Zip:     "06000",
			},
			ID: PlaceID("fr_06000_nice"),
		},
		ID: PlaceID("fr_06000_nice"),
	},
	{JSON: `{"country":"Kroatien","ciso":"hr","city":"Mali Lošinj","zip":"51553","id":"hr_51553_mali lošinj","boundingbox":[14.4693894,44.4778905,14.5315122,44.5260856]}`,
		STRUCT: Address{
			AddressFields: AddressFields{
				Country: Country{Country: "Kroatien", ID: CountryIDFromString("hr")},
				City:    "Mali Lošinj",
				Zip:     "51553",
			},
			ID:          PlaceID("hr_51553_mali lošinj"),
			BoundingBox: &Rect{14.4693894, 44.4778905, 14.5315122, 44.5260856},
		},
		ID: PlaceID("hr_51553_mali lošinj"),
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
		if d.STRUCT.BoundingBox != nil {
			assert.Equal(t, d.STRUCT.BoundingBox, a.BoundingBox)
		}
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
