package rest

import (
	"net/http"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/rest/views"
	"github.com/gorilla/mux"
)

type TimelineHandler struct {
	index library.DateIndex
	lib   library.PhotoLibrary
}

func NewTimelineHandler(index library.DateIndex, lib library.PhotoLibrary) *TimelineHandler {
	return &TimelineHandler{
		index: index,
		lib:   lib,
	}
}

func (dates *TimelineHandler) InitRoutes(r *mux.Router) {
	r.HandleFunc("/timeline/photos/{from}/{to}", dates.getTimelineForward).Methods("GET")
	r.HandleFunc("/timeline/index", dates.getTimelineIndex).Methods("GET")
}

type date string

func (dates *TimelineHandler) getTimelineForward(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	from := parseDateOrDefault(vars["from"], time.Time{})
	to := parseDateOrDefault(vars["to"], time.Now())
	ids, err := dates.index.FindRange(ctx, from, to)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	var photoViews []views.Photo
	for _, id := range ids {
		if p, err := dates.lib.Get(ctx, id); err == nil {
			photoViews = append(photoViews, views.PhotoFrom(p))
		}
	}
	respondWithJSON(w, http.StatusOK, photoViews)
}

func (dates *TimelineHandler) getTimelineIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	keys, err := dates.index.Keys(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, keys)
}

func parseDateOrDefault(date string, d time.Time) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return d
	}
	return t
}
