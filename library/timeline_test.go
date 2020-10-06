package library

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimelineAdd(t *testing.T) {
	data := []struct {
		dates []time.Time
		keys  []string

		years    []int
		months   []string
		expected []string
	}{
		{
			dates:    dates("2020-04-12", "2020-05-20", "2020-04-12"),
			keys:     []string{"one", "two", "three"},
			years:    []int{2020},
			months:   []string{"2020-04", "2020-05"},
			expected: []string{"one", "three", "two"},
		},
		{
			dates:    dates("2020-04-12", "2020-05-20", "2020-04-12", "2020-05-20"),
			keys:     []string{"one", "two", "three", "four"},
			years:    []int{2020},
			months:   []string{"2020-04", "2020-05"},
			expected: []string{"one", "three", "two", "four"},
		},
	}

	for _, d := range data {
		var line Timeline
		for i, date := range d.dates {
			line.Add(date, d.keys[i])
		}
		var result []string
		var years []int
		var months []string
		for _, y := range line.Years {
			years = append(years, y.Year)
			for _, m := range y.Months {
				months = append(months, fmt.Sprintf("%d-%02d", y.Year, m.Month))
				for _, day := range m.Days {
					result = append(result, day.Date)
				}
			}
		}
		assert.Equal(t, d.years, years)
		assert.Equal(t, d.months, months)
		assert.Equal(t, d.expected, result)
	}
}

func dates(in ...string) []time.Time {
	result := make([]time.Time, len(in))
	for i, s := range in {
		ts, err := time.Parse("2006-01-02", s)
		if err != nil {
			panic(err)
		}
		result[i] = ts
	}
	return result
}
