package rest

import (
	"net/http"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/rest/cursor"
	"bitbucket.org/kleinnic74/photos/rest/views"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type GeoHandler struct {
	index  library.GeoIndex
	photos library.PhotoLibrary
}

func NewGeoHandler(index library.GeoIndex, photos library.PhotoLibrary) *GeoHandler {
	return &GeoHandler{
		index:  index,
		photos: photos,
	}
}

func (g *GeoHandler) InitRoutes(r *mux.Router) {
	r.HandleFunc("/geo/photos/{country}/{zip}", g.getPhotos).Methods("GET")
	r.HandleFunc("/geo/photos/{country}", g.getPhotos).Methods("GET")
	r.HandleFunc("/geo/index", g.getGeoIndex).Methods("GET")
}

func (g *GeoHandler) getGeoIndex(w http.ResponseWriter, r *http.Request) {
	locations, err := g.index.Locations(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, locations)
}

func (g *GeoHandler) getPhotos(w http.ResponseWriter, r *http.Request) {
	log, ctx := logging.SubFrom(r.Context(), "geo")
	c := cursor.DecodeFromRequest(r)
	vars := mux.Vars(r)
	country, zip := vars["country"], vars["zip"]
	var photos []library.PhotoID
	var hasMore bool
	var err error
	if zip != "" {
		photos, hasMore, err = g.index.FindByPlacePaged(ctx, country, zip, c.Start, c.PageSize)
	} else {
		photos, hasMore, err = g.index.FindByCountryPaged(ctx, country, c.Start, c.PageSize)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	var v []views.Photo
	for _, p := range photos {
		if photo, err := g.photos.Get(ctx, p); err == nil {
			v = append(v, views.PhotoFrom(photo))
		} else {
			log.Warn("Unknown photo referenced in geoindex", zap.String("id", string(p)))
		}
	}
	respondWithJSON(w, http.StatusOK, cursor.PageFor(v, c, hasMore))
}
