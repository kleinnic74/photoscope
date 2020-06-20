package classification_test

import (
	"testing"
	"time"

	"bitbucket.org/kleinnic74/photos/classification"
)

type WithTimestamp time.Time

func TestNewDistanceMatrix(t *testing.T) {
	data := []classification.Timestamped{
		parseTime("2018-02-24T15:30:30Z"),
		parseTime("2018-01-13T16:30:00Z"),
		parseTime("2018-01-24T15:25:00Z"),
	}
	mat := classification.NewDistanceMatrix(data)
	if len(mat) != 3 {
		t.Fatalf("Expected matrix of 3x3, but got: %dx%d", len(mat), len(mat[0]))
	}
}

func (t WithTimestamp) Timestamp() time.Time {
	return time.Time(t)
}

func parseTime(value string) WithTimestamp {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return WithTimestamp(t)
}
