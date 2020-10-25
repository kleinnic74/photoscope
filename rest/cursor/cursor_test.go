package cursor_test

import (
	"testing"

	"bitbucket.org/kleinnic74/photos/rest/cursor"
	"github.com/stretchr/testify/assert"
)

const (
	start0Page20     = "eyJTdGFydCI6MCwiUGFnZVNpemUiOjIwfQ"
	start20Page20    = "eyJTdGFydCI6MjAsIlBhZ2VTaXplIjoyMH0"
	start3000Page100 = "eyJTdGFydCI6MzAwMCwiUGFnZVNpemUiOjEwMH0"
)

func TestEncodeCursor(t *testing.T) {
	data := []struct {
		Cursor   cursor.Cursor
		Expected string
	}{
		{
			Cursor:   cursor.Cursor{Start: 0, PageSize: 20},
			Expected: start0Page20,
		},
		{
			Cursor:   cursor.Cursor{Start: 20, PageSize: 20},
			Expected: start20Page20,
		},
		{
			Cursor:   cursor.Cursor{Start: 3000, PageSize: 100},
			Expected: start3000Page100,
		},
	}
	for i, d := range data {
		encoded := d.Cursor.Encode()
		if encoded != d.Expected {
			t.Errorf("#%d: Bad value for encoded cursor: expected %s, got %s", i, d.Expected, encoded)
		}
	}
}

func TestDecodeCursor(t *testing.T) {
	data := []struct {
		Encoded  string
		Expected cursor.Cursor
		Next     cursor.Cursor
		hasNext  bool
	}{
		{
			Encoded:  start0Page20,
			Expected: cursor.Cursor{Start: 0, PageSize: 20},
			Next:     cursor.Cursor{Start: 20, PageSize: 20},
			hasNext:  true,
		},
		{
			Encoded:  start20Page20,
			Expected: cursor.Cursor{Start: 20, PageSize: 20},
			Next:     cursor.Cursor{Start: 40, PageSize: 20},
			hasNext:  true,
		},
		{
			Encoded:  "",
			Expected: cursor.Cursor{Start: 0, PageSize: 33},
			Next:     cursor.Cursor{Start: 33, PageSize: 33},
			hasNext:  true,
		},
		{
			Encoded:  start3000Page100,
			Expected: cursor.Cursor{Start: 3000, PageSize: 100},
			Next:     cursor.Cursor{Start: 3100, PageSize: 100},
			hasNext:  true,
		},
	}
	for i, d := range data {
		actual := cursor.Cursor{PageSize: 33}
		if err := cursor.DecodeFromString(d.Encoded, &actual); err != nil {
			t.Fatalf("Failed to decode cursor: %s", err)
		}
		assert.Equal(t, d.Expected, actual, "%d: bad cursor value", i)
		next, hasNext := actual.Next()
		if hasNext != d.hasNext {
			t.Fatalf("%d: Bad subsequent cursor state: expected %t, got %t", i, d.hasNext, hasNext)
		}
		assert.Equal(t, d.Next, next, "%d: bad NEXT cursor", i)
	}
}
