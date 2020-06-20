package domain_test

import (
	"testing"

	"bitbucket.org/kleinnic74/photos/domain"
)

func TestNewEvent(t *testing.T) {
	event := domain.NewEvent(2018, 2, 24)
	if event.Date != "2018-02-24" {
		t.Fatalf("Bad date value for event: expeted %s, got %s", "2018-02-24", event.Date)
	}
}
