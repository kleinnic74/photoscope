package rest

import (
	"context"
	"net/http"
	"time"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func WithMiddleWares(handler http.Handler, name string) http.Handler {
	return cors(addRequestID(logRequest(handler, name)))
}

type responseWrapper struct {
	writer http.ResponseWriter
	status int
}

func (w *responseWrapper) Header() http.Header {
	return w.writer.Header()
}

func (w *responseWrapper) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

func (w *responseWrapper) WriteHeader(status int) {
	w.status = status
	w.writer.WriteHeader(status)
}

func logRequest(f http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log := logging.From(r.Context())
		rid := r.Context().Value(requestIDKey).(string)
		ctx := logging.Context(r.Context(), log.With(zap.String("request", rid)))
		wrapper := responseWrapper{
			writer: w,
		}
		defer func() {
			log := logging.From(r.Context()).Named(name)
			log.Debug("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
				zap.String("request", rid),
				zap.Int("status", wrapper.status))
		}()
		f.ServeHTTP(&wrapper, r.WithContext(ctx))
	})
}

type requestIDKeyType int

const requestIDKey = requestIDKeyType(0)

func addRequestID(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
			r.Header.Set("X-Request-ID", rid)
		}
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		f.ServeHTTP(w, r.WithContext(ctx))
	})
}

func cors(h http.Handler) http.Handler {
	if consts.IsDevMode() {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			h.ServeHTTP(w, r)
		})
	} else {
		return h
	}
}
