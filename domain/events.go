package domain

import "time"

type Event struct {
	Date  string
	start time.Time
	end   time.Time
}

func NewEvent(year, month, day int) *Event {
	start := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	return &Event{
		Date:  start.Format("2006-01-02"),
		start: start,
		end:   end,
	}
}
