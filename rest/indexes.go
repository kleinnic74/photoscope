package rest

import (
	"errors"
	"net/http"

	"bitbucket.org/kleinnic74/photos/index"
	"bitbucket.org/kleinnic74/photos/rest/cursor"
	"github.com/gorilla/mux"
)

type NoSuchIndex string

var MissingIndexNameError = errors.New("Path parameter 'name' missing")

func (e NoSuchIndex) Error() string {
	return string(e)
}

type Indexes struct {
	indexes  *index.Indexer
	migrator *index.MigrationCoordinator
}

func NewIndexes(indexer *index.Indexer, migrator *index.MigrationCoordinator) *Indexes {
	return &Indexes{indexes: indexer, migrator: migrator}
}

func (i *Indexes) Init(router *mux.Router) {
	router.Path("/indexes/state/{name}").Methods(http.MethodGet).HandlerFunc(i.getIndexStatus)
	router.Path("/indexes/elements").Methods(http.MethodGet).HandlerFunc(i.getIndexPhotoStatus)
	router.Path("/indexes").Methods(http.MethodGet).HandlerFunc(i.getIndexes)
}

func (i *Indexes) getIndexes(w http.ResponseWriter, r *http.Request) {
	c := cursor.DecodeFromRequest(r)
	indexes := i.indexes.GetIndexes()
	Respond(r).WithJSON(w, http.StatusOK, cursor.PageFor(indexes, c, false))
}

func (i *Indexes) getIndexStatus(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	status, found := i.migrator.GetIndexStatus(index.Name(name))
	responder := Respond(r)
	if !found {
		responder.WithError(w, http.StatusNotFound, NoSuchIndex(name))
		return
	}
	responder.WithJSON(w, http.StatusOK, cursor.Unpaged(status))
}

func (i *Indexes) getIndexPhotoStatus(w http.ResponseWriter, r *http.Request) {
	responder := Respond(r)
	indexingStatus, err := i.indexes.GetElementStatus(r.Context())
	if err != nil {
		responder.WithError(w, http.StatusInternalServerError, err)
		return
	}
	responder.WithJSON(w, http.StatusOK, indexingStatus)
}
