package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
)

type DebugHandler struct{}

func (d DebugHandler) InitRoutes(router *mux.Router) {
	// mux introspection
	router.HandleFunc("/debug/mux", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Context-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintln(rw, "<html><head><title>Endpoints</title></head><body>")
		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			t, err := route.GetPathTemplate()
			if err != nil {
				return err
			}
			fmt.Fprintln(rw, fmt.Sprintf("<div><a href=\"%s\">%s</a></div>", t, t))
			return nil
		})
		fmt.Fprintln(rw, "</body></html>")
	}).Methods(("GET"))

	// pprof routes
	router.HandleFunc("/debug/pprof/", pprof.Index).Methods("GET")
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
	router.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
}
