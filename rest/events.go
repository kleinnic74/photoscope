package rest

import (
	"errors"
	"net/http"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/library/boltstore"
	"bitbucket.org/kleinnic74/photos/rest/cursor"
	"bitbucket.org/kleinnic74/photos/rest/views"
	"github.com/gorilla/mux"
)

var (
	errorNoID = errors.New("Missing 'id'")
)

type EventsHandler struct {
	events *boltstore.EventIndex
	lib    library.PhotoLibrary
}

func NewEventsHandler(events *boltstore.EventIndex, lib library.PhotoLibrary) *EventsHandler {
	return &EventsHandler{events, lib}
}

func (h *EventsHandler) InitRoutes(r *mux.Router) {
	r.HandleFunc("/events", h.listEvents).Methods(http.MethodGet)
	r.HandleFunc("/events/{id}", h.photosForEvent).Methods(http.MethodGet)
}

func (h *EventsHandler) listEvents(w http.ResponseWriter, r *http.Request) {
	page := cursor.DecodeFromRequest(r)
	e, hasMore, err := h.events.FindPaged(r.Context(), page.Start, page.PageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, cursor.PageFor(e, page, hasMore))
}

func (h *EventsHandler) photosForEvent(w http.ResponseWriter, r *http.Request) {
	eventID, ok := mux.Vars(r)["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, errorNoID)
		return
	}
	page := cursor.DecodeFromRequest(r)
	photoIDs, hasMore, err := h.events.FindPhotosPaged(r.Context(), eventID, page.Start, page.PageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	v := make([]views.Photo, len(photoIDs))
	for i, p := range photoIDs {
		if photo, err := h.lib.Get(r.Context(), p); err != nil {
			v[i] = views.PhotoFrom(photo)
		}
	}
	respondWithJSON(w, http.StatusOK, cursor.PageFor(v, page, hasMore))
}
