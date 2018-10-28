package rest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
)

var (
	a   *App
	lib library.PhotoLibrary
)

func TestMain(m *testing.M) {
	lib = newPhotoLib()
	a = NewApp(lib)

	code := m.Run()
	os.Exit(code)
}

func TestGetAll(t *testing.T) {
	req, _ := http.NewRequest("GET", "/photos", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.router.ServeHTTP(rr, req)
	return rr

}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Bad response code: expected %d, got %d", expected, actual)
	}
}

type testLib struct {
	photos []domain.Photo
}

func (lib *testLib) Add(p domain.Photo) error {
	lib.photos = append(lib.photos, p)
	return nil
}

func (lib *testLib) FindAll() []domain.Photo {
	return lib.photos
}

func (lib *testLib) FindAllPaged(start, max uint) []domain.Photo {
	return lib.photos
}

func (lib *testLib) Find(start, end time.Time) []domain.Photo {
	return lib.photos
}

func (lib *testLib) Get(id string) (domain.Photo, error) {
	for _, p := range lib.photos {
		if p.ID() == id {
			return p, nil
		}
	}
	return nil, library.NotFound(id)
}

func newPhotoLib() library.PhotoLibrary {
	return &testLib{photos: make([]domain.Photo, 0)}
}
