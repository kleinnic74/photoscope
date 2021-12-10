package swarm

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"

	"bitbucket.org/kleinnic74/photos/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type remoteThumber struct {
	baseURL *url.URL
	client  *http.Client

	thumbFormat domain.Format

	requestCount prometheus.Counter
	errorCount   prometheus.Counter
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

		requestCount: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem:   "thumberremote",
			Name:        "requests_total",
			Help:        "Total number of remote thumb requests sent",
			ConstLabels: prometheus.Labels{"url": baseURL},
		}),
		errorCount: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem:   "thumberremote",
			Name:        "errors_total",
			Help:        "Total number of erroneous remote thumb requests",
			ConstLabels: prometheus.Labels{"url": baseURL},
		}),
	}, nil
}

func (t *remoteThumber) CreateThumb(in io.Reader, f domain.Format, o domain.Orientation, size domain.ThumbSize) (thumb image.Image, err error) {
	defer func() {
		t.requestCount.Inc()
		if err != nil {
			t.errorCount.Inc()
		}
	}()

	endpoint := fmt.Sprintf("%s/%s/%s", t.baseURL.String(), t.thumbFormat.ID(), size.Name)
	var r *http.Request
	if r, err = http.NewRequest(http.MethodPost, endpoint, in); err != nil {
		return
	}
	r.Header.Set("Content-Type", f.Mime())
	var resp *http.Response
	if resp, err = t.client.Do(r); err != nil {
		return
	}
	defer resp.Body.Close()

	thumb, err = f.Decode(resp.Body)
	return
}
