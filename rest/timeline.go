package rest

import (
	"net/http"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/rest/cursor"
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
	r.HandleFunc("/timeline/photos", dates.getTimelineForward).Methods("GET")
	r.HandleFunc("/timeline/index", dates.getTimelineIndex).Methods("GET")
}

type date string

func (dates *TimelineHandler) getTimelineForward(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c := cursor.DecodeFromRequest(r)
	from := parseDateOrDefault(r.FormValue("from"), time.Time{})
	to := parseDateOrDefault(r.FormValue("to"), time.Now())
	ids, hasMore, err := dates.index.FindRangePaged(ctx, from, to, c.Start, c.PageSize)
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
	respondWithJSON(w, http.StatusOK, cursor.PageFor(photoViews, c, hasMore))
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
