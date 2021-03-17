package logging

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"bitbucket.org/kleinnic74/photos/consts"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	userAgent string = "photoscope (" + consts.GitRepo + "@" + consts.GitRepo + ")"

	endpointFormat = "https://logs-01.loggly.com/bulk/%s/tag/bulk/"
)

type logglySink struct {
	url    string
	client *http.Client

	threshold int
	done      chan struct{}
	q         chan []byte
}

func NewLogglyEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(func() zapcore.EncoderConfig {
		cfg := zap.NewProductionEncoderConfig()
		cfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		cfg.TimeKey = "timestamp"
		return cfg
	}())
}

func NewLogglySink(token string) zap.Sink {
	sink := &logglySink{
		url:       fmt.Sprintf(endpointFormat, token),
		client:    &http.Client{},
		threshold: 4096,
		done:      make(chan struct{}),
		q:         make(chan []byte, 10),
	}
	go sink.drain()
	return sink
}

func (s *logglySink) Write(p []byte) (int, error) {
	cpy := make([]byte, len(p))
	copy(cpy, p)
	s.q <- cpy
	return len(p), nil
}

func (s *logglySink) Sync() error {
	return nil // nothing to do
}

func (s *logglySink) Close() error {
	close(s.done)
	return nil
}

func (s *logglySink) drain() {
	var buffer bytes.Buffer
	fmt.Fprintf(os.Stderr, "Forwarding logs to loggly: %s\n", s.url)
	for {
		select {
		case b := <-s.q:
			buffer.Write(b)
			if buffer.Len() > s.threshold {
				s.push(buffer.Bytes())
				buffer.Reset()
			}
		case <-s.done:
			fmt.Fprintf(os.Stderr, "Loggly: terminating\n")
			return
		}
	}
}

func (s *logglySink) push(data []byte) {
	post, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loggly: failed to create HTTP POST request: %s\n", err)
		return
	}
	post.Header.Add("User-Agent", userAgent)
	post.Header.Add("Content-Type", "application/json")
	post.Header.Add("Content-Length", strconv.Itoa(len(data)))
	r, err := s.client.Do(post)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loggly: failed to create HTTP POST request: %s\n", err)
		return
	}
	if r.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Loggly: failed to create HTTP POST request: %s\n", err)
		return
	}
	return
}
