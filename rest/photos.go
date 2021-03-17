package rest

import (
	"fmt"
	"net/http"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/rest/cursor"
	"bitbucket.org/kleinnic74/photos/rest/views"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var (
	jpg = domain.MustFormatForExt("jpg")
)

// App is the REST API that can be used as an http.HandlerFunc
type App struct {
	router *mux.Router
	lib    library.PhotoLibrary
}

// NewApp creates a new instance of the REST application
// as an http.HandlerFunc
func NewApp(lib library.PhotoLibrary) (a *App) {
	return &App{lib: lib}

}

func (a *App) InitRoutes(r *mux.Router) {
	r.HandleFunc("/photos/{id}/view", a.getPhotoImage).Methods("GET").Name("/photos/{id}/view")
	r.HandleFunc("/photos/{id}/thumb", a.getThumb).Methods("GET").Name("/photos/{id}/thumb")
	r.HandleFunc("/photos/{id}", a.getPhoto).Methods("GET").Name("/photos/{id}")
	r.HandleFunc("/photos", a.getPhotos).Methods("GET").Name("/photos")
}

func (a *App) route(path string, f http.HandlerFunc) *mux.Route {
	return a.router.HandleFunc(path, f)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *App) getPhotos(w http.ResponseWriter, r *http.Request) {
	c := cursor.DecodeFromRequest(r)
	responder := Respond(r)
	photos, hasMore, err := a.lib.FindAllPaged(r.Context(), c.Start, c.PageSize, c.Order)
	if err != nil {
		responder.WithError(w, http.StatusInternalServerError, err)
		return
	}
	logging.From(r.Context()).Named("http").Info("/photos",
		zap.Bool("hasMore", hasMore), zap.Int("start", c.Start), zap.Int("page", c.PageSize))
	photoViews := make([]views.Photo, len(photos))
	for i, p := range photos {
		photoViews[i] = views.PhotoFrom(p)
	}
	responder.WithJSON(w, http.StatusOK, cursor.PageFor(photoViews, c, hasMore))
}

func (a *App) getPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := library.PhotoID(vars["id"])
	responder := Respond(r)
	photo, err := a.lib.Get(r.Context(), id)
	if photo == nil && err == nil {
		responder.WithError(w, http.StatusNotFound, fmt.Errorf("No photo with id %s", id))
		return
	}
	if err != nil {
		responder.WithError(w, http.StatusInternalServerError, err)
		return
	}
	responder.WithJSON(w, http.StatusOK, photo)
}

func (a *App) getPhotoImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := library.PhotoID(vars["id"])
	responder := Respond(r)
	binary, photo, err := a.lib.OpenContent(r.Context(), id)
	if binary == nil && err == nil {
		responder.WithError(w, http.StatusNotFound, fmt.Errorf("No photo with id %s", id))
		return
	}
	if err != nil {
		responder.WithError(w, http.StatusInternalServerError, err)
		return
	}
	defer binary.Close()
	respondWithBinary(w, photo.Format.Mime(), photo.Size, binary)
}

func (a *App) getThumb(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := library.PhotoID(vars["id"])
	responder := Respond(r)
	thumb, format, err := a.lib.OpenThumb(r.Context(), id, domain.Small)
	if err != nil {
		switch err.(type) {
		case library.ErrNotFound:
			responder.WithError(w, http.StatusNotFound, fmt.Errorf("No photo with id %s", id))
		case domain.ErrThumbsNotSupported:
			responder.WithError(w, http.StatusNotImplemented, err)
		default:
			logging.From(r.Context()).Error("Internal error", zap.Error(err))
			responder.WithError(w, http.StatusInternalServerError, err)
		}
		return
	}
	defer thumb.Close()
	respondWithBinary(w, format.Mime(), 0, thumb)
	return
}
