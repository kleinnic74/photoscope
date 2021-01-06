package library

import (
	"bytes"
	"context"
	"io"
	"time"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/domain"
)

// PhotoID is the unique identifier of a Photo
type PhotoID string

// OrderedID is an ID whose values can be sorted based on time
type OrderedID []byte

type ExtendedPhotoID struct {
	ID     PhotoID   `json:"id"`
	SortID OrderedID `json:"sortId,omitempty"`
}

type AsendingOrderedIDs []OrderedID

func (a AsendingOrderedIDs) Len() int           { return len(a) }
func (a AsendingOrderedIDs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AsendingOrderedIDs) Less(i, j int) bool { return bytes.Compare(a[i], a[j]) == -1 }

// PhotoLibrary represents the operations on a library of photos
type PhotoLibrary interface {
	Add(ctx context.Context, photo domain.Photo, content io.Reader) error
	Get(ctx context.Context, id PhotoID) (*Photo, error)
	FindAll(ctx context.Context, order consts.SortOrder) ([]*Photo, error)
	FindAllPaged(ctx context.Context, start, maxCount int, order consts.SortOrder) ([]*Photo, bool, error)
	Find(ctx context.Context, start, end time.Time, order consts.SortOrder) ([]*Photo, error)

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

func (h BinaryHash) String() string {
	return string(h)
}

// Store represents a persistent storage of photo meta-data
type Store interface {
	Exists(hash BinaryHash) (PhotoID, bool)
	Add(*Photo) error
	Update(p *Photo) error
	Get(id PhotoID) (*Photo, error)
	FindAll(order consts.SortOrder) ([]*Photo, error)
	FindAllPaged(start, maxCount int, order consts.SortOrder) ([]*Photo, bool, error)
	Find(start, end OrderedID, order consts.SortOrder) ([]*Photo, error)
}

// ClosableStore is a Store that can be closed
type ClosableStore interface {
	Store

	Close()
}
