package library

import (
	"encoding/json"
	"fmt"
	"testing"

	"bitbucket.org/kleinnic74/photos/domain"
	"github.com/stretchr/testify/assert"
)

func TestPhotoJSONMarshalling(t *testing.T) {
	data := []struct {
		json  string
		photo Photo
	}{
		{`{"schema":1,"id":"123","or":4,"format":"jpg"}`, Photo{schema: 1, ID: "123", Orientation: 4, Format: domain.MustFormatForExt("jpg")}},
	}
	for i, d := range data {
		name := fmt.Sprintf("#%d:", i)
		t.Run(name, func(t *testing.T) {
			unmarshall(t, d.json, d.photo)
		})
	}
}

func unmarshall(t *testing.T, data string, expected Photo) {
	var actual Photo
	if err := json.Unmarshal([]byte(data), &actual); err != nil {
		t.Fatalf("Failed to decode JSON: %s", err)
	}
	assert.Equal(t, expected, actual)
}
