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

var photoData = []struct {
	json  string
	photo Photo
}{
	{fmt.Sprintf(`{"id":"123","sortId":"AQIDBA==","schema":%d,"format":"jpg","dateUN":1573148532000000000,"or":4}`, currentSchema),
		Photo{schema: currentSchema,
			ExtendedPhotoID: ExtendedPhotoID{ID: "123", SortID: []byte{1, 2, 3, 4}},
			Orientation:     4,
			Format:          domain.MustFormatForExt("jpg"),
			DateTaken:       time.Date(2019, 11, 07, 17, 42, 12, 0, time.UTC),
		}},
}

func TestPhotoJSONUnmarshal(t *testing.T) {
	for i, d := range photoData {
		name := fmt.Sprintf("#%d:", i)
		t.Run(name, func(t *testing.T) {
			var actual Photo
			if err := json.Unmarshal([]byte(d.json), &actual); err != nil {
				t.Fatalf("Failed to decode JSON: %s", err)
			}
			assert.Equal(t, d.photo, actual)
		})
	}
}

func TestPhotoJSONMarshal(t *testing.T) {
	for i, d := range photoData {
		name := fmt.Sprintf("#%d:", i)
		t.Run(name, func(t *testing.T) {
			encoded, err := json.Marshal(&d.photo)
			if err != nil {
				t.Fatalf("Failed to marshal to JSON: %s", err)
			}
			assert.Equal(t, d.json, string(encoded))
		})
	}
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
