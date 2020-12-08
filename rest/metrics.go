package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHandler struct {
	handler http.Handler
}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		handler: promhttp.Handler(),
	}
}

func (m *MetricsHandler) InitRoutes(r *mux.Router) {
	r.Handle("/metrics", m.handler).Methods("GET")
}
