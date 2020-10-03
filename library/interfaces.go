package library

import (
	"context"
	"io"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
)

// PhotoID is the unique identifier of a Photo
type PhotoID string

// PhotoLibrary represents the operations on a library of photos
type PhotoLibrary interface {
	Add(ctx context.Context, photo domain.Photo, content io.Reader) error
	Get(ctx context.Context, id PhotoID) (*Photo, error)
	FindAll(ctx context.Context) ([]*Photo, error)
	FindAllPaged(ctx context.Context, start, maxCount int) ([]*Photo, bool, error)
	Find(ctx context.Context, start, end time.Time) ([]*Photo, error)

	OpenContent(ctx context.Context, id PhotoID) (io.ReadCloser, *Photo, error)
	OpenThumb(ctx context.Context, id PhotoID, size domain.ThumbSize) (io.ReadCloser, domain.Format, error)
}

type PhotoIndex interface {
	Add(ctx context.Context, photo *Photo) error
}

// BinaryHash is the hash of the binary content of the photo file, this
// is used to detect duplicate files
type BinaryHash string

func (h BinaryHash) Bytes() []byte {
	return []byte(h)
}

// Store represents a persistent storage of photo meta-data
type Store interface {
	Exists(hash BinaryHash) (PhotoID, bool)
	Add(*Photo) error
	Get(id PhotoID) (*Photo, error)
	FindAll() ([]*Photo, error)
	FindAllPaged(start, maxCount int) ([]*Photo, bool, error)
	Find(start, end time.Time) ([]*Photo, error)
}

// ClosableStore is a Store that can be closed
type ClosableStore interface {
	Store

	Close()
}
