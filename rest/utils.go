package rest

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"

	"bitbucket.org/kleinnic74/photos/domain"
)

var ()

type Responder interface {
	WithJSON(http.ResponseWriter, int, interface{})
	WithError(http.ResponseWriter, int, error)
}

type encoderFunc func(*json.Encoder) *json.Encoder

type responder struct {
	encoderOptions encoderFunc
}

var (
	pretty  Responder
	compact Responder
)

func init() {
	pretty = responder{func(encoder *json.Encoder) *json.Encoder {
		encoder.SetIndent("", "  ")
		return encoder
	}}
	compact = responder{func(e *json.Encoder) *json.Encoder { return e }}
}

func Respond(r *http.Request) Responder {
	if r.URL.Query().Get("pretty") == "true" {
		return pretty
	}
	return compact
}

func (r responder) WithError(w http.ResponseWriter, status int, err error) {
	r.WithJSON(w, status, map[string]string{"error": err.Error()})
}

func (r responder) WithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := r.encoderOptions(json.NewEncoder(w))
	encoder.Encode(payload)
}

func respondWithBinary(w http.ResponseWriter, mime string, size int64, data io.Reader) {
	w.Header().Set("Content-Type", mime)
	if size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	w.WriteHeader(http.StatusOK)
	io.Copy(w, data)
}

func respondWithImage(w http.ResponseWriter, format domain.Format, image image.Image) {
	w.Header().Set("Content-Type", format.Mime())
	w.WriteHeader(http.StatusOK)
	format.Encode(image, w)
}

type simplePayload struct {
	Data interface{} `json:"data,omitempty"`
}
