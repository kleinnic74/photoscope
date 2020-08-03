package rest

import (
	"context"
	"errors"
	"image"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
	"github.com/gorilla/mux"
)

var (
	a   *App
	lib library.PhotoLibrary
)

func TestGetAll(t *testing.T) {
	lib = newPhotoLib()
	a = NewApp(lib)

	router := mux.NewRouter()
	a.InitRoutes(router)

	req, _ := http.NewRequest("GET", "/photos", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	response := rr.Result()

	checkResponseCode(t, http.StatusOK, response)
}

func checkResponseCode(t *testing.T, expected int, response *http.Response) {
	if expected != response.StatusCode {
		t.Fatalf("Bad response code: expected %d, got %d (%s)", expected, response.StatusCode, response.Status)
	}
}

type testLib struct {
	photos []domain.Photo
}

func (lib *testLib) Add(ctx context.Context, p domain.Photo) error {
	lib.photos = append(lib.photos, p)
	return nil
}

func (lib *testLib) FindAll(ctx context.Context) []domain.Photo {
	return lib.photos
}

func (lib *testLib) FindAllPaged(ctx context.Context, start, max uint) []domain.Photo {
	return lib.photos
}

func (lib *testLib) Find(ctx context.Context, start, end time.Time) []domain.Photo {
	return lib.photos
}

func (lib *testLib) Get(ctx context.Context, id string) (domain.Photo, error) {
	for _, p := range lib.photos {
		if p.ID() == id {
			return p, nil
		}
	}
	return nil, library.NotFound(id)
}

func (lib *testLib) Thumb(ctx context.Context, id string, size domain.ThumbSize) (image.Image, domain.Format, error) {
	return nil, nil, errors.New("Not implemented")
}

func newPhotoLib() library.PhotoLibrary {
	return &testLib{photos: make([]domain.Photo, 0)}
}
