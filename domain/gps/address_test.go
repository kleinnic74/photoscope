package gps

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var addressData = []struct {
	JSON   string
	STRUCT Address
}{
	{JSON: `{"country":"France","ciso":"FR","city":"Nice","zip":"06000"}`,
		STRUCT: Address{Country: Country{Country: "France", Code: "FR"},
			Place: Place{City: "Nice", Zip: "06000"},
		},
	},
}

func TestUnmarshalAddress(t *testing.T) {
	for i, d := range addressData {
		var a Address
		if err := json.Unmarshal([]byte(d.JSON), &a); err != nil {
			t.Fatalf("#%d: error while unmarshalling JSON: %s", i, err)
		}
		assert.Equal(t, d.STRUCT, a)
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
