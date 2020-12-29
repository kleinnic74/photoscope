package rest

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/kleinnic74/photos/events"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/gorilla/mux"
)

type EventHandler struct {
	events *events.Stream
}

func NewEventHandler(stream *events.Stream) *EventHandler {
	return &EventHandler{
		events: stream,
	}
}

func (e *EventHandler) InitRoutes(router *mux.Router) {
	router.HandleFunc("/eventstream", e.listen).Methods("GET")
}

func (e *EventHandler) listen(w http.ResponseWriter, r *http.Request) {
	logger := logging.From(r.Context())
	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Warn("HTTP Flusher not supported")
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	e.events.Listen(r.Context(), func(event events.Event) {
		if err := json.NewEncoder(w).Encode(event); err != nil {
			return
		}
		flusher.Flush()
	})
}
