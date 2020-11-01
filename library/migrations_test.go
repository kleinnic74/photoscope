package library

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrations(t *testing.T) {
	migrations := instanceMigrations()
	migrations.Register(1, InstanceFunc(func(ctx context.Context, p Photo, _ ReaderFunc) (Photo, error) {
		p.Orientation = 1
		return p, nil
	}))
	data := []struct {
		src          Photo
		dst          Photo
		expectChange bool
	}{
		{src: Photo{schema: 0, Orientation: 0},
			dst: Photo{schema: currentSchema, Orientation: 1, Hash: "E5d6oXltiEgafGIYfZLv7g=="}, expectChange: true},
		{src: Photo{schema: 1, Orientation: 0, Hash: "E5d6oXltiEgafGIYfZLv7g=="},
			dst: Photo{schema: currentSchema, Orientation: 0, Hash: "E5d6oXltiEgafGIYfZLv7g=="}, expectChange: true},
	}
	for i, d := range data {
		ctx := context.Background()
		result, err := migrations.Apply(ctx, d.src, func() (io.ReadCloser, error) {
			return os.Open("../domain/testdata/orientation/portrait_3.jpg")
		})
		if err != nil {
			t.Fatalf("Error while applying migrations: %s", err)
		}
		changed := result != d.src
		if changed != d.expectChange {
			t.Errorf("#%d: Unexpected change status, expected %t, got %t", i, d.expectChange, changed)
		}
		assert.Equal(t, d.dst, result, "Photo #%d", i)
	}
}
