package rest

import (
	"net/http"

	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/gorilla/mux"
)

type logsHandler struct{}

func NewLogsHandler() logsHandler {
	return logsHandler{}
}

func (l logsHandler) InitRoutes(r *mux.Router) {
	r.Handle("/logs", l).Methods("GET")
}

func (l logsHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	logging.Dump(w, true)
}
