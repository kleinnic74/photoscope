package library

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
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
)

type NewPhotoCallback func(ctx context.Context, p *Photo) error

// BasicPhotoLibrary is a library storing photos on the filesystem
type BasicPhotoLibrary struct {
	basedir  string
	photodir string
	dirMode  os.FileMode
	db       ClosableStore

	thumbdir    string
	thumbFormat domain.Format

	callbacks []NewPhotoCallback
}

// ReaderFunc is a function providing an io.ReadCloser
type ReaderFunc func() (io.ReadCloser, error)

func wrap(in io.ReadCloser) ReaderFunc {
	return func() (io.ReadCloser, error) {
		return in, nil
	}
}

type ErrNotFound PhotoID

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("No photo with id %s", string(e))
}

// NotFound Error to indicate that the photo with the given id does not exist
func NotFound(id PhotoID) error {
	return ErrNotFound(id)
}

// PhotoAlreadyExists Error to indicate that the photo with the given id already exists
func PhotoAlreadyExists(id PhotoID) error {
	return fmt.Errorf("Photo already exists: id=%s", id)
}

// PhotoFileAlreadyExists indicates that a given photo file already exists in the library
func PhotoFileAlreadyExists(path string) error {
	return fmt.Errorf("Photo already exists at path=%s", path)
}

// NewBasicPhotoLibrary creates a new photo library at the given directory using the given meta-data store provider function
func NewBasicPhotoLibrary(basedir string, store ClosableStore) (*BasicPhotoLibrary, error) {
	absdir, err := filepath.Abs(basedir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absdir, defaultDirMode); err != nil {
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
		db:       store,

		thumbdir:    thumbsDir,
		thumbFormat: domain.MustFormatForExt("jpg"),
	}, nil
}

func (lib *BasicPhotoLibrary) AddCallback(callback NewPhotoCallback) {
	lib.callbacks = append(lib.callbacks, callback)
}

// Add adds a photo to this library. If the given photo already exists, then
// an error of type PhotoAlreadyExists is returned
func (lib *BasicPhotoLibrary) Add(ctx context.Context, photo domain.Photo, content io.Reader) error {
	ctx = logging.Context(ctx, logging.From(ctx).Named("library").With(zap.String("source", photo.Name())))
	targetDir, name, id := canonicalizeFilename(photo)
	content, hash, err := loadContent(content)
	if err != nil {
		return err
	}
	if dup, exists := lib.db.Exists(hash); exists {
		return PhotoAlreadyExists(dup)
	}
	size, err := lib.addPhotoFile(ctx, content, lib.photodir, targetDir, name)
	if err != nil {
		return err
	}
	path := filepath.Join(targetDir, name)
	p := &Photo{
		Path:      path,
		ID:        id,
		DateTaken: photo.DateTaken().UTC(),
		Location:  photo.Location(),
		Format:    photo.Format(),
		Size:      size,
	}
	logging.From(ctx).Info("Added", zap.String("photo", string(id)), zap.Any("location", p.Location))
	if err := lib.db.Add(p); err != nil {
		return err
	}
	for _, cb := range lib.callbacks {
		cb(ctx, p)
	}
	return nil
}

// Get returns the photo with the given ID
func (lib *BasicPhotoLibrary) Get(ctx context.Context, id PhotoID) (*Photo, error) {
	return lib.db.Get(id)
}

// FindAll returns all photos from the underlying store
func (lib *BasicPhotoLibrary) FindAll(ctx context.Context) ([]*Photo, error) {
	return lib.db.FindAll()
}

// FindAllPaged returns maximum maxCount photos from the underlying store starting
// at start index
func (lib *BasicPhotoLibrary) FindAllPaged(ctx context.Context, start, maxCount int) ([]*Photo, bool, error) {
	return lib.db.FindAllPaged(start, maxCount)
}

// Find returns all photos stored in this library that have been taken between
// the given start and end times
func (lib *BasicPhotoLibrary) Find(ctx context.Context, start, end time.Time) ([]*Photo, error) {
	return lib.db.Find(start, end)
}

// OpenContent returns an io.ReadCloser on the content of the photo with the given ID.
// The caller is responsible to close the reader
func (lib *BasicPhotoLibrary) OpenContent(ctx context.Context, id PhotoID) (io.ReadCloser, *Photo, error) {
	p, err := lib.db.Get(id)
	if err != nil {
		return nil, nil, err
	}
	reader, err := lib.openPhoto(p.Path)
	return reader, p, err
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

func (lib *BasicPhotoLibrary) addPhotoFile(ctx context.Context, in io.Reader, basedir, targetDir, targetName string) (int64, error) {
	pathInLib := filepath.Join(basedir, targetDir, targetName)
	log, ctx := logging.FromWithFields(ctx, zap.String("dest", pathInLib))
	if _, err := os.Stat(pathInLib); err == nil {
		// File already exists
		return 0, PhotoFileAlreadyExists(filepath.Join(targetDir, targetName))
	}
	log.Debug("Adding...")
	err := lib.createDirectory(ctx, basedir, targetDir)
	if err != nil {
		return 0, err
	}
	// Does not exist yet, copy
	out, err := os.Create(pathInLib)
	if err != nil {
		log.Error("Could not add photo", zap.Error(err))
		return 0, err
	}
	defer out.Close()
	size, err := io.Copy(out, in)
	if err != nil {
		log.Error("Could not copy photo to library", zap.Error(err))
		return 0, err
	}
	log.Info("Added photo")
	return size, nil
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

func (lib *BasicPhotoLibrary) OpenThumb(ctx context.Context, id PhotoID, size domain.ThumbSize) (io.ReadCloser, domain.Format, error) {
	logger, ctx := logging.FromWithNameAndFields(ctx, "library", zap.String("photo", string(id)))

	photo, err := lib.Get(ctx, id)
	if err != nil {
		// Photo does not exist
		return nil, nil, err
	}
	dir := filepath.Join(lib.thumbdir, string(photo.ID))
	path := filepath.Join(dir, size.Name+".jpg")
	if _, err := os.Stat(path); err != nil {
		// Thumb does not exist yet
		logger = logger.With(zap.String("thumb", path))
		start := time.Now()
		logger.Debug("Creating thumbnail")
		if _, err := os.Stat(dir); err != nil {
			if err = os.MkdirAll(dir, defaultDirMode); err != nil {
				return nil, nil, err
			}
		}

		baseImage, err := lib.openPhoto(photo.Path)
		if err != nil {
			logger.Error("Failed to open image content", zap.Error(err))
			return nil, nil, err
		}
		thumb, err := photo.Format.Thumb(baseImage, domain.Small)
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
	}
	// Thumb exists
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return f, lib.thumbFormat, nil
}

func canonicalizeFilename(photo domain.Photo) (dir, filename string, id PhotoID) {
	dir = photo.DateTaken().Format("2006/01/02")
	filename = fmt.Sprintf("%s.%s", photo.ID(), photo.Format().ID())
	h := mmh3.New128()
	h.Write([]byte(photo.DateTaken().Format(time.RFC3339)))
	h.Write([]byte(strings.ToLower(filename)))
	id = PhotoID(fmt.Sprintf("%x", h.Sum(nil)))
	return
}

func loadContent(in io.Reader) (io.Reader, BinaryHash, error) {
	h := mmh3.New128()
	in = io.TeeReader(in, h)
	content := new(bytes.Buffer)
	if _, err := io.Copy(content, in); err != nil {
		return nil, "", err
	}
	hash := BinaryHash(base64.StdEncoding.EncodeToString(h.Sum(nil)))
	return content, hash, nil
}
