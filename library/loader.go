package library

import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"

	"github.com/reusee/mmh3"
)

type Loader string

func NewLoader(tmpDir string) Loader {
	return Loader(tmpDir)
}

func (l Loader) LoadMediaObject(name string, in io.Reader) (MediaObject, error) {
	h := mmh3.New128()
	switch i := in.(type) {
	case io.ReadSeeker:
		// we can read the content and then reset the reader
		if _, err := io.Copy(h, in); err != nil {
			return nil, err
		}
		if _, err := i.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		return mediaReadSeeker{in: i, hash: BinaryHash(base64.StdEncoding.EncodeToString(h.Sum(nil)))}, nil
	default:
		// Canot seek in the stream, must temporarily store the content
		staging := filepath.Join(string(l), name)
		tmp, err := os.Create(staging)
		if err != nil {
			return nil, err
		}
		defer tmp.Close()
		in = io.TeeReader(in, h)
		if _, err := io.Copy(tmp, in); err != nil {
			defer func() { os.Remove(staging) }()
			return nil, err
		}
		return stagedMediaObject{path: staging, hash: BinaryHash(base64.StdEncoding.EncodeToString(h.Sum(nil)))}, nil
	}

}

type ContentHandlerFunc func(io.Reader) error

type MediaObject interface {
	Hash() BinaryHash
	Cleanup()
	ProcessContent(ContentHandlerFunc) error
}

type mediaReadSeeker struct {
	hash BinaryHash
	in   io.Reader
}

func (m mediaReadSeeker) Hash() BinaryHash {
	return m.hash
}

func (m mediaReadSeeker) Cleanup() {
	// Nothing to do here
}

func (m mediaReadSeeker) ProcessContent(f ContentHandlerFunc) error {
	return f(m.in)
}

type stagedMediaObject struct {
	hash BinaryHash
	path string
}

func (m stagedMediaObject) Hash() BinaryHash {
	return m.hash
}

func (m stagedMediaObject) Cleanup() {
	os.Remove(m.path)
}

func (m stagedMediaObject) ProcessContent(f ContentHandlerFunc) error {
	in, err := os.Open(m.path)
	if err != nil {
		return err
	}
	defer in.Close()
	return f(in)
}
