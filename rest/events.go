package rest

import (
	"net/http"

	"bitbucket.org/kleinnic74/photos/library"
	"github.com/gorilla/mux"
)

type Classifiers struct {
	lib library.PhotoLibrary
}

func NewClassifiers(lib library.PhotoLibrary) *Classifiers {
	return &Classifiers{lib}
}

func (c *Classifiers) InitRoutes(r *mux.Router) {
	r.HandleFunc("/events", c.distanceMatrix).Methods(http.MethodGet)
}

func (c *Classifiers) distanceMatrix(w http.ResponseWriter, r *http.Request) {

}
