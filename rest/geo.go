package rest

import (
	"net/http"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/geocoding"
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
	r.HandleFunc("/geo/photos/byplace/{placeID}", g.getPhotosByPlace).Methods("GET")
	r.HandleFunc("/geo/photos/bycountry/{countryID}", g.getPhotosByCountry).Methods("GET")
	r.HandleFunc("/geo/index", g.getGeoIndex).Methods("GET")
}

func (g *GeoHandler) getGeoIndex(w http.ResponseWriter, r *http.Request) {
	responder := Respond(r)
	locations, err := g.index.Locations(r.Context())
	if err != nil {
		responder.WithError(w, http.StatusInternalServerError, err)
		return
	}
	responder.WithJSON(w, http.StatusOK, locations)
}

func (g *GeoHandler) getPhotosByPlace(w http.ResponseWriter, r *http.Request) {
	log, ctx := logging.SubFrom(r.Context(), "geo")
	c := cursor.DecodeFromRequest(r)
	vars := mux.Vars(r)
	placeID := gps.PlaceID(vars["placeID"])
	responder := Respond(r)
	photos, hasMore, err := g.index.FindByPlacePaged(ctx, placeID, c.Start, c.PageSize)
	if err != nil {
		responder.WithError(w, http.StatusInternalServerError, err)
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
	responder.WithJSON(w, http.StatusOK, cursor.PageFor(v, c, hasMore))
}

func (g *GeoHandler) getPhotosByCountry(w http.ResponseWriter, r *http.Request) {
	log, ctx := logging.SubFrom(r.Context(), "geo")
	c := cursor.DecodeFromRequest(r)
	vars := mux.Vars(r)
	responder := Respond(r)
	countryID := gps.CountryID(vars["countryID"])
	photos, hasMore, err := g.index.FindByCountryPaged(ctx, countryID, c.Start, c.PageSize)
	if err != nil {
		responder.WithError(w, http.StatusInternalServerError, err)
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
	responder.WithJSON(w, http.StatusOK, cursor.PageFor(v, c, hasMore))
}

type GeoCacheHandler struct {
	cache *geocoding.Cache
}

func NewGeoCacheHandler(cache *geocoding.Cache) *GeoCacheHandler {
	return &GeoCacheHandler{cache: cache}
}

func (g *GeoCacheHandler) InitRoutes(r *mux.Router) {
	r.HandleFunc("/geo/cache/stats", g.getGeoCache).Methods("GET")
	r.HandleFunc("/geo/cache/view", g.renderSvg).Methods("GET")
}

func (g *GeoCacheHandler) getGeoCache(w http.ResponseWriter, r *http.Request) {
	Respond(r).WithJSON(w, http.StatusOK, g.cache.DumpStats())
}

func (g *GeoCacheHandler) renderSvg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	renderer := geocoding.NewGeoView(w)
	g.cache.Visit(renderer)
}
