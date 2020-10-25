package swarm

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"

	"bitbucket.org/kleinnic74/photos/domain"
)

type remoteThumber struct {
	baseURL *url.URL
	client  *http.Client

	thumbFormat domain.Format
}

func NewRemoteThumber(baseURL string, thumbFormat domain.Format) (domain.Thumber, error) {
	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &remoteThumber{
		baseURL:     endpoint,
		client:      &http.Client{},
		thumbFormat: thumbFormat,
	}, nil
}

func (t *remoteThumber) CreateThumb(in io.Reader, f domain.Format, o domain.Orientation, size domain.ThumbSize) (image.Image, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", t.baseURL.String(), t.thumbFormat.ID(), size.Name)
	r, err := http.NewRequest(http.MethodPost, endpoint, in)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", f.Mime())
	resp, err := t.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	thumb, err := f.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return thumb, nil
}
