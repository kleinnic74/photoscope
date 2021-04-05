package rest

import (
	"fmt"
	"net/http"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ThumberAPI struct {
	thumber domain.Thumber
}

func NewThumberAPI(t domain.Thumber) *ThumberAPI {
	return &ThumberAPI{t}
}

func (t *ThumberAPI) InitRoutes(r *mux.Router) {
	r.HandleFunc("/thumb/{format}/{size}", t.createThumb).Methods(http.MethodPost).Name("/thumb")
}

func (t *ThumberAPI) createThumb(w http.ResponseWriter, r *http.Request) {
	f := mux.Vars(r)["format"]
	format, found := domain.FormatForExt(f)
	if !found {
		Respond(r).WithError(w, http.StatusBadRequest, fmt.Errorf("No such format '%s'", f))
		return
	}
	s := mux.Vars(r)["size"]
	size, found := domain.ThumbSizes[s]
	if !found {
		Respond(r).WithError(w, http.StatusBadRequest, fmt.Errorf("Invalid thumb size '%s'", s))
		return
	}
	contentType := r.Header.Get("Content-Type")
	inputFormat, found := domain.FormatForMime(contentType)
	if !found {
		Respond(r).WithError(w, http.StatusNotImplemented, fmt.Errorf("Input format '%s' not suported", contentType))
		return
	}
	thumb, err := t.thumber.CreateThumb(r.Body, inputFormat, domain.NormalOrientation, size)
	if err != nil {
		Respond(r).WithError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusOK)

	if err := format.Encode(thumb, w); err != nil {
		logging.From(r.Context()).Warn("Failed to encode thumb", zap.Error(err))
	}
}
