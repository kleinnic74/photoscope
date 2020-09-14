//go:generate go run generator.go

package embed

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type Resource struct {
	data []byte
	Type string
}

func inferType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	t := mime.TypeByExtension("." + ext)
	if t == "" {
		return "application/octect-stream"
	}
	return t
}

func (r Resource) Size() int {
	return len(r.data)
}

func (r Resource) Open() io.Reader {
	return bytes.NewReader(r.data)
}

type embedded map[string]Resource

func newEmbedded() embedded {
	return make(map[string]Resource)
}

func (e embedded) Add(path string, content []byte) {
	e[path] = Resource{data: content, Type: inferType(path)}
}

func (e embedded) Get(path string) ([]byte, error) {
	r, found := e[path]
	if !found {
		return nil, os.ErrNotExist
	}
	return r.data, nil
}

func (e embedded) Open(path string) (io.ReadCloser, error) {
	data, err := e.Get(path)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(data)), nil
}

func (e embedded) GetResource(path string) (Resource, error) {
	r, found := e[path]
	if !found {
		return Resource{}, os.ErrNotExist
	}
	return r, nil
}

var e = newEmbedded()

func Add(path string, context []byte) {
	e.Add(path, context)
}

func Open(path string) (io.ReadCloser, error) {
	return e.Open(path)
}

func Get(path string) ([]byte, error) {
	return e.Get(path)
}

func GetResource(path string) (Resource, error) {
	return e.GetResource(path)
}
