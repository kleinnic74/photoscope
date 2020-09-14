package wdav

import (
	"net/http"

	"golang.org/x/net/webdav"
)

func NewWebDavHandler(tmpdir string, callback UploadedFunc) (http.Handler, error) {
	wdav, err := NewWebDavAdapter(tmpdir, callback)
	if err != nil {
		return nil, err
	}
	return &webdav.Handler{
		Prefix:     "/dav/",
		LockSystem: webdav.NewMemLS(),
		FileSystem: wdav,
	}, nil
}
