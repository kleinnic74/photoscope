package library

import (
	"context"
	"time"

	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type DateInfo struct {
	Date  string `json:"key"`
	Count int
}

type Timeline struct {
	Years []*Year `json:"years"`
}

func (timeline *Timeline) Add(t time.Time, key string) {
	year := t.Year()
	for _, y := range timeline.Years {
		if y.Year == year {
			y.Add(t, key)
			return
		}
	}
	newYear := &Year{Year: year}
	newYear.Add(t, key)
	timeline.Years = append(timeline.Years, newYear)
}

type Year struct {
	Year   int      `json:"year"`
	Months []*Month `json:"months"`
}

func (year *Year) Add(t time.Time, key string) {
	month := t.Month()
	for _, m := range year.Months {
		if month == m.Month {
			m.Add(t, key)
		}
	}
	newMonth := &Month{Month: month}
	newMonth.Add(t, key)
	year.Months = append(year.Months, newMonth)
}

type Month struct {
	Month time.Month  `json:"month"`
	Days  []*DateInfo `json:"days"`
}

func (m *Month) Add(t time.Time, key string) {
	for _, d := range m.Days {
		if d.Date == key {
			return
		}
	}
	newDay := &DateInfo{Date: key}
	m.Days = append(m.Days, newDay)
}

type DateIndex interface {
	Keys(context.Context) (Timeline, error)
	Add(context.Context, *Photo) error
	FindRange(context.Context, time.Time, time.Time) ([]PhotoID, error)
}

type GeoIndex interface {
	Has(context.Context, PhotoID) bool
	Get(context.Context, PhotoID) (*gps.Address, bool, error)
	Update(context.Context, PhotoID, *gps.Address) error
}
