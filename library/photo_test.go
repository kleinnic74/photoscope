package library

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

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

func TestByteOrderOfId(t *testing.T) {
	var data = []struct {
		ts string
		id PhotoID
	}{
		{"2016-12-26T17:09:11Z", PhotoID("cde")},
		{"2016-12-31T17:46:11Z", PhotoID("cde")},
		{"2017-02-24T15:22:18Z", PhotoID("abc")},
	}
	for k, v := range data[1:] {
		tK, _ := time.Parse(time.RFC3339, v.ts)
		tKm1, _ := time.Parse(time.RFC3339, data[k].ts)
		if tK.Before(tKm1) {
			t.Fatalf("Bad time stamp order, ts[%d] is before ts[%d]", k+1, k)
		}
		idK := orderedIDOf(tK, v.id)
		idKm1 := orderedIDOf(tKm1, data[k].id)
		if bytes.Compare(idK, idKm1) <= 0 {
			t.Errorf("Bad byte order, id[%d] is lower than id[%d] (%s <= %s)", k+1, k, tK, tKm1)
		}
	}
}
