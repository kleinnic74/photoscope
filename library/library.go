package library

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/reusee/mmh3"
	"go.uber.org/zap"
)

const (
	defaultDirMode = 0755
	dbName         = "photos.db"
)

// PhotoLibrary represents the operations on a library of photos
type PhotoLibrary interface {
	Add(ctx context.Context, photo domain.Photo) error
	Get(ctx context.Context, id string) (domain.Photo, error)
	FindAll(ctx context.Context) []domain.Photo
	FindAllPaged(ctx context.Context, start, maxCount uint) []domain.Photo
	Find(ctx context.Context, start, end time.Time) []domain.Photo
	Thumb(ctx context.Context, id string, size domain.ThumbSize) (image.Image, domain.Format, error)
}

// Store represents a persistent storage of photo meta-data
type Store interface {
	Exists(dateTaken time.Time, id string) bool
	Add(*Photo) error
	Get(id string) (*Photo, error)
	FindAll() []*Photo
	FindAllPaged(start, maxCount uint) []*Photo
	Find(start, end time.Time) []*Photo
}

// ClosableStore is a Store that can be closed
type ClosableStore interface {
	Store

	Close()
}

// StoreBuilder function to create a store at the given directory with the given name
type StoreBuilder func(string, string) (ClosableStore, error)

// BasicPhotoLibrary is a library storing photos on the filesystem
type BasicPhotoLibrary struct {
	basedir  string
	photodir string
	dirMode  os.FileMode
	db       ClosableStore

	thumbdir    string
	thumbFormat domain.Format
}

// ReaderFunc is a function providing an io.ReadCloser
type ReaderFunc func() (io.ReadCloser, error)

func wrap(in io.ReadCloser) ReaderFunc {
	return func() (io.ReadCloser, error) {
		return in, nil
	}
}

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("No photo with id %s", string(e))
}

// NotFound Error to indicate that the photo with the given id does not exist
func NotFound(id string) error {
	return ErrNotFound(id)
}

// PhotoAlreadyExists Error to indicate that the photo with the given id already exists
func PhotoAlreadyExists(id string) error {
	return fmt.Errorf("Photo already exists: id=%s", id)
}

// NewBasicPhotoLibrary creates a new photo library at the given directory using the given meta-data store provider function
func NewBasicPhotoLibrary(basedir string, store StoreBuilder) (*BasicPhotoLibrary, error) {
	absdir, err := filepath.Abs(basedir)
	if err != nil {
		return nil, err
	}
	if info, err := os.Stat(absdir); err != nil {
		err = os.MkdirAll(absdir, defaultDirMode)
		if err != nil {
			return nil, err
		}
	} else if !info.IsDir() {
		err = fmt.Errorf("%s exists but is not a directory", absdir)
		return nil, err
	}
	db, err := store(absdir, dbName)
	if err != nil {
		return nil, err
	}
	photosDir := filepath.Join(absdir, "photos")
	if err := os.MkdirAll(photosDir, defaultDirMode); err != nil {
		return nil, err
	}
	thumbsDir := filepath.Join(absdir, "thumbs")
	if err := os.MkdirAll(thumbsDir, defaultDirMode); err != nil {
		return nil, err
	}
	return &BasicPhotoLibrary{
		basedir:  absdir,
		photodir: photosDir,
		dirMode:  defaultDirMode,
		db:       db,

		thumbdir:    thumbsDir,
		thumbFormat: domain.MustFormatForExt("jpg"),
	}, nil
}

// Add adds a photo to this library. If the given photo already exists, then
// an error of type PhotoAlreadyExists is returned
func (lib *BasicPhotoLibrary) Add(ctx context.Context, photo domain.Photo) error {
	ctx = logging.Context(ctx, logging.From(ctx).Named("library").With(zap.String("source", photo.Name())))
	targetDir, name, id := canonicalizeFilename(photo)
	if lib.db.Exists(photo.DateTaken().UTC(), id) {
		return PhotoAlreadyExists(id)
	}
	content, err := photo.Content()
	if err != nil {
		return err
	}
	if err := lib.addPhotoFile(ctx, content, lib.photodir, targetDir, name); err != nil {
		return err
	}
	p := &Photo{
		lib:       lib,
		path:      filepath.Join(targetDir, name),
		id:        id,
		dateTaken: photo.DateTaken().UTC(),
		location:  *photo.Location(),
		format:    photo.Format(),
	}
	logging.From(ctx).Info("Added", zap.String("photo", id), zap.Any("location", p.location))
	return lib.db.Add(p)
}

// Get returns the photo with the given ID
func (lib *BasicPhotoLibrary) Get(ctx context.Context, id string) (domain.Photo, error) {
	p, err := lib.db.Get(id)
	if err != nil {
		return nil, err
	}
	p.lib = lib
	return p, err
}

// FindAll returns all photos from the underlying store
func (lib *BasicPhotoLibrary) FindAll(ctx context.Context) []domain.Photo {
	var result = make([]domain.Photo, 0)
	for _, p := range lib.db.FindAll() {
		p.lib = lib
		result = append(result, p)
	}
	return result
}

// FindAllPaged returns maximum maxCount photos from the underlying store starting
// at start index
func (lib *BasicPhotoLibrary) FindAllPaged(ctx context.Context, start, maxCount uint) []domain.Photo {
	var result = make([]domain.Photo, 0)
	for _, p := range lib.db.FindAllPaged(start, maxCount) {
		p.lib = lib
		result = append(result, p)
	}
	return result
}

// Find returns all photos stored in this library that have been taken between
// the given start and end times
func (lib *BasicPhotoLibrary) Find(ctx context.Context, start, end time.Time) []domain.Photo {
	var result = make([]domain.Photo, 0)
	for _, p := range lib.db.Find(start, end) {
		p.lib = lib
		result = append(result, p)
	}
	return result
}

func (lib *BasicPhotoLibrary) createDirectory(ctx context.Context, basedir, dir string) error {
	fullpath := filepath.Join(basedir, dir)
	if info, err := os.Stat(fullpath); err != nil {
		logging.From(ctx).Info("Creating directory", zap.String("dir", fullpath))
		return os.MkdirAll(fullpath, lib.dirMode)
	} else if !info.IsDir() {
		return fmt.Errorf("Error: %s exists but is not a directory", fullpath)
	} else {
		return nil
	}
}

func (lib *BasicPhotoLibrary) addPhotoFile(ctx context.Context, in io.ReadCloser, basedir, targetDir, targetName string) error {
	pathInLib := filepath.Join(basedir, targetDir, targetName)
	log, ctx := logging.FromWithFields(ctx, zap.String("dest", pathInLib))
	if _, err := os.Stat(pathInLib); err == nil {
		// File already exists
		return PhotoAlreadyExists(filepath.Join(targetDir, targetName))
	}
	log.Debug("Adding...")
	err := lib.createDirectory(ctx, basedir, targetDir)
	if err != nil {
		return err
	}
	// Does not exist yet, copy
	out, err := os.Create(pathInLib)
	if err != nil {
		log.Error("Could not add photo", zap.Error(err))
		return err
	}
	defer out.Close()
	defer in.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		log.Error("Could not copy photo to library", zap.Error(err))
		return err
	}
	log.Info("Added photo")
	return nil
}

func (lib *BasicPhotoLibrary) openPhoto(path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(lib.photodir, path))
}

func (lib *BasicPhotoLibrary) fileSizeOf(path string) int64 {
	info, err := os.Stat(filepath.Join(lib.photodir, path))
	if err != nil {
		return -1
	}
	return info.Size()
}

func (lib *BasicPhotoLibrary) Thumb(ctx context.Context, id string, size domain.ThumbSize) (image.Image, domain.Format, error) {
	ctx = logging.Context(ctx, logging.From(ctx).Named("library").With(zap.String("photo", id)))
	logger := logging.From(ctx)

	photo, err := lib.Get(ctx, id)
	if err != nil {
		// Photo does not exist
		return nil, nil, err
	}
	dir := filepath.Join(lib.thumbdir, photo.ID())
	path := filepath.Join(dir, size.Name+".jpg")
	if _, err := os.Stat(path); err != nil {
		// Thumb does not exist yet
		logger = logger.With(zap.String("photo", id), zap.String("thumb", path))
		start := time.Now()
		logger.Debug("Creating thumbnail")
		if _, err := os.Stat(dir); err != nil {
			if err = os.MkdirAll(dir, defaultDirMode); err != nil {
				return nil, nil, err
			}
		}

		baseImage, err := photo.Content()
		if err != nil {
			logger.Error("Failed to open image content", zap.Error(err))
			return nil, nil, err
		}
		thumb, err := photo.Format().Thumb(baseImage, domain.Small)
		if err != nil {
			logger.Error("Failed to created thumb", zap.Error(err))
			return nil, nil, err
		}
		out, err := os.Create(path)
		if err != nil {
			logger.Error("Failed to save thumb", zap.Error(err))
			return nil, nil, err
		}
		defer out.Close()
		lib.thumbFormat.Encode(thumb, out)
		logger.Info("Created thumb", zap.Duration("duration", time.Since(start)))
		return thumb, lib.thumbFormat, nil
	}
	// Thumb exists
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	img, err := lib.thumbFormat.Decode(f)
	return img, lib.thumbFormat, err
}

func canonicalizeFilename(photo domain.Photo) (dir, filename, id string) {
	dir = photo.DateTaken().Format("2006/01/02")
	filename = fmt.Sprintf("%s.%s", photo.ID(), photo.Format().ID())
	h := mmh3.New128()
	h.Write([]byte(photo.DateTaken().Format(time.RFC3339)))
	h.Write([]byte(strings.ToLower(filename)))
	id = fmt.Sprintf("%x", h.Sum(nil))
	return
}
