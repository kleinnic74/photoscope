package rest

import (
	"net/http"

	"bitbucket.org/kleinnic74/photos/index"
	"bitbucket.org/kleinnic74/photos/rest/cursor"
	"github.com/gorilla/mux"
)

type NoSuchIndex string

func (e NoSuchIndex) Error() string {
	return string(e)
}

type Indexes struct {
	indexes *index.MigrationCoordinator
}

func NewIndexes(indexer *index.MigrationCoordinator) *Indexes {
	return &Indexes{indexes: indexer}
}

func (i *Indexes) Init(router *mux.Router) {
	router.Path("/indexes/{name}").Methods(http.MethodGet).HandlerFunc(i.getIndexStatus)
	router.Path("/indexes").Methods(http.MethodGet).HandlerFunc(i.getIndexes)
}

func (i *Indexes) getIndexes(w http.ResponseWriter, r *http.Request) {
	c := cursor.DecodeFromRequest(r)
	indexes := i.indexes.GetIndexes()
	respondWithJSON(w, http.StatusOK, cursor.PageFor(indexes, c, false))
}

func (i *Indexes) getIndexStatus(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	status, found := i.indexes.GetIndexStatus(index.Name(name))
	if !found {
		respondWithError(w, http.StatusNotFound, NoSuchIndex(name))
		return
	}
	respondWithJSON(w, http.StatusOK, cursor.Unpaged(status))
}
