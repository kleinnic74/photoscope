package rest

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
	"github.com/gorilla/mux"
)

var (
	jpg = domain.MustFormatForExt("jpg")
)

// App is the REST API that can be used as an http.HandlerFunc
type App struct {
	router *mux.Router
	lib    library.PhotoLibrary
}

type middleware func(http.HandlerFunc) http.HandlerFunc

func logging() middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() {
				log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
			}()
			f(w, r)
		}
	}
}

func chain(f http.HandlerFunc, middlewares ...middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

// NewApp creates a new instance of the REST application
// as an http.HandlerFunc
func NewApp(lib library.PhotoLibrary) (a *App) {
	a = &App{router: mux.NewRouter(), lib: lib}
	a.route("/photos/{id}", a.getPhoto).Methods("GET")
	a.route("/photos/{id}/view", a.getPhotoImage).Methods("GET")
	a.route("/photos", a.getPhotos).Methods("GET")
	a.route("/thumb/{id}", a.getThumb).Methods("GET")
	return a
}

func (a *App) route(path string, f http.HandlerFunc) *mux.Route {
	return a.router.HandleFunc(path, chain(f, logging()))
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *App) getPhotos(w http.ResponseWriter, r *http.Request) {
	photos := a.lib.FindAll()
	respondWithJSON(w, http.StatusOK, photos)
}

func (a *App) getPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	photo, err := a.lib.Get(id)
	if photo == nil && err == nil {
		respondWithError(w, http.StatusNotFound, fmt.Errorf("No photo with id %s", id))
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, photo)
}

func (a *App) getPhotoImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	photo, err := a.lib.Get(id)
	if photo == nil && err == nil {
		respondWithError(w, http.StatusNotFound, fmt.Errorf("No photo with id %s", id))
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	imgReader, err := photo.Content()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	defer imgReader.Close()
	respondWithBinary(w, photo.Format().Mime, photo.SizeInBytes(), imgReader)
}

func (a *App) getThumb(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	photo, err := a.lib.Get(id)
	if photo == nil && err == nil {
		respondWithError(w, http.StatusNotFound, fmt.Errorf("No photo with id %s", id))
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	t, err := photo.Thumb(domain.Small)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithImage(w, jpg, t)
	return
}

func respondWithError(w http.ResponseWriter, status int, err error) {
	respondWithJSON(w, status, map[string]string{"error": err.Error()})
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondWithBinary(w http.ResponseWriter, mime string, size int64, data io.Reader) {
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, data)
}

func respondWithImage(w http.ResponseWriter, format *domain.Format, image image.Image) {
	w.Header().Set("Content-Type", format.Mime)
	w.WriteHeader(http.StatusOK)
	format.Encode(image, w)
}
