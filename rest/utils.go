package rest

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"

	"bitbucket.org/kleinnic74/photos/domain"
)

func respondWithError(w http.ResponseWriter, status int, err error) {
	respondWithJSON(w, status, map[string]string{"error": err.Error()})
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
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
