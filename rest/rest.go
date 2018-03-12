package rest

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/kleinnic74/photos/library"
	"github.com/gorilla/mux"
)

type App struct {
	router *mux.Router
	lib    library.PhotoLibrary
}

func NewApp(lib library.PhotoLibrary) (a *App) {
	a = &App{router: mux.NewRouter(), lib: lib}
	a.router.HandleFunc("/photos", a.getPhotos).Methods("GET")
	return a

}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *App) getPhotos(w http.ResponseWriter, r *http.Request) {
	photos := a.lib.FindAll()
	respondWithJSON(w, http.StatusOK, photos)
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
