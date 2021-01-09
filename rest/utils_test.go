package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponder(t *testing.T) {
	data := []struct {
		URL  string
		JSON string
	}{
		{"http://host/path?pretty=false", `{"id":"1234"}` + "\n"},
		{"http://host/path?pretty=true", `{
  "id": "1234"
}` + "\n"},
	}
	for _, d := range data {
		t.Run(d.URL, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, d.URL, nil)
			resp := httptest.NewRecorder()
			Respond(r).WithJSON(resp, http.StatusOK, struct {
				ID string `json:"id"`
			}{"1234"})
			body, err := ioutil.ReadAll(resp.Result().Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %s", err)
			}
			assert.Equal(t, d.JSON, string(body))
		})
	}
}
