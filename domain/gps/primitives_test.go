package gps_test

import (
	"testing"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/stretchr/testify/assert"
)

func TestPointsAndRect(t *testing.T) {
	inside := gps.Point{0, 0}
	outside := gps.Point{2, 0}
	bounds := gps.Rect{-1, -1, 1, 2}
	assert.True(t, inside.In(bounds))
	assert.False(t, outside.In(bounds))
}
