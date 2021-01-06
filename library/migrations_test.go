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
		{src: Photo{schema: 0, ExtendedPhotoID: ExtendedPhotoID{ID: "1234"}, Orientation: 0},
			dst: Photo{schema: currentSchema,
				ExtendedPhotoID: ExtendedPhotoID{
					ID:     "1234",
					SortID: []byte{0x30, 0x30, 0x30, 0x31, 0x2d, 0x30, 0x31, 0x2d, 0x30, 0x31, 0x54, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x5a, 0xc3, 0x5d, 0x1c, 0x72},
				}, Orientation: 1, Hash: "E5d6oXltiEgafGIYfZLv7g=="}, expectChange: true},
		{src: Photo{schema: 1, Orientation: 0, Hash: "E5d6oXltiEgafGIYfZLv7g=="},
			dst: Photo{schema: currentSchema,
				ExtendedPhotoID: ExtendedPhotoID{
					SortID: []byte{0x30, 0x30, 0x30, 0x31, 0x2d, 0x30, 0x31, 0x2d, 0x30, 0x31, 0x54, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x5a, 0x0, 0x0, 0x0, 0x0},
				}, Orientation: 0, Hash: "E5d6oXltiEgafGIYfZLv7g=="}, expectChange: true},
	}
	for i, d := range data {
		ctx := context.Background()
		result, err := migrations.Apply(ctx, d.src, func() (io.ReadCloser, error) {
			return os.Open("../domain/testdata/orientation/portrait_3.jpg")
		})
		if err != nil {
			t.Fatalf("Error while applying migrations: %s", err)
		}
		changed := result.schema != d.src.schema
		if changed != d.expectChange {
			t.Errorf("#%d: Unexpected change status, expected %t, got %t", i, d.expectChange, changed)
		}
		assert.Equal(t, d.dst, result, "Photo #%d", i)
	}
}
