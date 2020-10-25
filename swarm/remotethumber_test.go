package swarm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"bitbucket.org/kleinnic74/photos/domain"
	"github.com/stretchr/testify/assert"
)

func TestCreateThumber(t *testing.T) {
	srcImg := resource(t, "testdata/Canon_40D.jpg")
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write(srcImg)
	}))
	defer server.Close()

	baseURL := fmt.Sprintf("%s/%s", server.URL, "base")
	thumber, err := NewRemoteThumber(baseURL, domain.JPEG)
	if err != nil {
		t.Fatalf("Failed to create thumber: %s", err)
	}
	jpg, _ := domain.FormatForExt("jpg")
	thumb, err := thumber.CreateThumb(bytes.NewBuffer(srcImg), jpg, domain.Orientation(1), domain.Small)
	if err != nil {
		t.Fatalf("Error while retrieving thumb: %s", err)
	}
	assert.Equal(t, "/base/jpg/S", requestedPath)
	assert.NotNil(t, thumb, "Expected a thumb in return")
}

func resource(t *testing.T, path string) (out []byte) {
	in, err := os.Open(path)
	if err != nil {
		t.Fatalf("Cannot open resource %s: %s", path, err)
	}
	defer in.Close()
	out, err = ioutil.ReadAll(in)
	return
}
