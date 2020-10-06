package boltstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestDateIndexFindDates(t *testing.T) {
	data := []struct {
		input    []time.Time
		expected []string
	}{
		{
			input:    times("2020-04-12T12:30:24Z", "2020-05-09T07:08:09Z", "2020-04-12T08:45:00Z"),
			expected: []string{"2020-04-12", "2020-05-09"},
		},
		{
			input:    times("2020-04-12T12:30:24Z", "2019-05-09T07:08:09Z", "2018-04-12T08:45:00Z"),
			expected: []string{"2018-04-12", "2019-05-09", "2020-04-12"},
		},
	}
	for i, d := range data {
		runTestWithBoltDB(t, func(t *testing.T, db *bolt.DB) {
			dateindex, err := NewDateIndex(db)
			if err != nil {
				t.Fatalf("Failed to create DateIndex: %s", err)
			}
			for _, ts := range d.input {
				photo := library.Photo{
					ID:        library.PhotoID(fmt.Sprintf("%d", i)),
					DateTaken: ts,
				}
				dateindex.Add(context.Background(), &photo)
			}
			timeline, err := dateindex.FindDates(context.Background())
			if err != nil {
				t.Fatalf("#%d: error while fetching dates: %s", i, err)
			}
			assert.Equal(t, d.expected, dates(timeline...))
		})
	}
}

func times(in ...string) (result []time.Time) {
	result = make([]time.Time, len(in))
	for i, s := range in {
		ts, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		result[i] = ts
	}
	return
}

func dates(in ...time.Time) (result []string) {
	result = make([]string, len(in))
	for i, t := range in {
		result[i] = t.Format("2006-01-02")
	}
	return
}
