package library

import (
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"github.com/reusee/mmh3"
	log "github.com/sirupsen/logrus"
)

const (
	defaultDirMode = 0755
	dbName         = "photos.db"
)

// PhotoLibrary represents the operations on a library of photos
type PhotoLibrary interface {
	Add(photo domain.Photo) error
	Get(id string) (domain.Photo, error)
	FindAll() []domain.Photo
	FindAllPaged(start, maxCount uint) []domain.Photo
	Find(start, end time.Time) []domain.Photo
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
	thumbFormat *domain.Format
}

// ReaderFunc is a function providing an io.ReadCloser
type ReaderFunc func() (io.ReadCloser, error)

func wrap(in io.ReadCloser) ReaderFunc {
	return func() (io.ReadCloser, error) {
		return in, nil
	}
}

// NotFound Error to indicate that the photo with the given id does not exist
func NotFound(id string) error {
	return fmt.Errorf("No photo with id %s", id)
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
	var lib = &BasicPhotoLibrary{
		basedir:  absdir,
		photodir: filepath.Join(absdir, "photos"),
		dirMode:  defaultDirMode,
		db:       db,

		thumbdir:    filepath.Join(absdir, "thumbs"),
		thumbFormat: domain.MustFormatForExt("jpg"),
	}
	return lib, nil
}

// Add adds a photo to this library. If the given photo already exists, then
// an error of type PhotoAlreadyExists is returned
func (lib *BasicPhotoLibrary) Add(photo domain.Photo) error {
	targetDir, name, id := canonicalizeFilename(photo)
	if lib.db.Exists(photo.DateTaken().UTC(), id) {
		return PhotoAlreadyExists(id)
	}
	if err := lib.addPhotoFile(photo.Content, lib.photodir, targetDir, name); err != nil {
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
	return lib.db.Add(p)
}

// Get returns the photo with the given ID
func (lib *BasicPhotoLibrary) Get(id string) (domain.Photo, error) {
	p, err := lib.db.Get(id)
	if err != nil {
		return nil, err
	}
	p.lib = lib
	return p, err
}

// FindAll returns all photos from the underlying store
func (lib *BasicPhotoLibrary) FindAll() []domain.Photo {
	var result = make([]domain.Photo, 0)
	for _, p := range lib.db.FindAll() {
		p.lib = lib
		result = append(result, p)
	}
	return result
}

// FindAllPaged returns maximum maxCount photos from the underlying store starting
// at start index
func (lib *BasicPhotoLibrary) FindAllPaged(start, maxCount uint) []domain.Photo {
	var result = make([]domain.Photo, 0)
	for _, p := range lib.db.FindAllPaged(start, maxCount) {
		p.lib = lib
		result = append(result, p)
	}
	return result
}

// Find returns all photos stored in this library that have been taken between
// the given start and end times
func (lib *BasicPhotoLibrary) Find(start, end time.Time) []domain.Photo {
	var result = make([]domain.Photo, 0)
	for _, p := range lib.db.Find(start, end) {
		p.lib = lib
		result = append(result, p)
	}
	return result
}

func (lib *BasicPhotoLibrary) createDirectory(basedir, dir string) error {
	fullpath := filepath.Join(basedir, dir)
	if info, err := os.Stat(fullpath); err != nil {
		return os.MkdirAll(fullpath, lib.dirMode)
	} else if !info.IsDir() {
		return fmt.Errorf("Error: %s exists but is not a directory", fullpath)
	} else {
		return nil
	}
}

func (lib *BasicPhotoLibrary) addPhotoFile(in ReaderFunc, basedir, targetDir, targetName string) error {
	pathInLib := filepath.Join(basedir, targetDir, targetName)
	if _, err := os.Stat(pathInLib); err == nil {
		// File already exists
		return PhotoAlreadyExists(filepath.Join(targetDir, targetName))
	}
	log.Infof("Adding file: %s", pathInLib)
	err := lib.createDirectory(basedir, targetDir)
	if err != nil {
		return err
	}
	// Does not exist yet, copy in the background
	go func() {
		out, err := os.Create(pathInLib)
		if out != nil {
			defer out.Close()
		}
		if err != nil {
			log.Errorf("Could not add photo '%s': %s", pathInLib, err)
			return
		}
		content, err := in()
		if content != nil {
			defer content.Close()
		}
		if err != nil {
			log.Errorf("Could not read photo content: %s", err)
			return
		}
		_, err = io.Copy(out, content)
		if err != nil {
			log.Errorf("Could not copy photo to library '%s': %s", pathInLib, err)
			return
		}
	}()
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

func (lib *BasicPhotoLibrary) openThumb(id string, size domain.ThumbSize) (image.Image, error) {
	path := filepath.Join(lib.thumbdir, id, size.Name+".jpg")
	if _, err := os.Stat(path); err != nil {
		photo, err := lib.Get(id)
		if err != nil {
			return nil, err
		}
		ch := make(chan image.Image)
		errch := make(chan error)
		go func() {
			defer close(ch)
			defer close(errch)
			log.Infof("Creating thumbnail for %s at %s...", id, path)
			src, err := photo.Image()
			if err != nil {
				log.Errorf("Could not read image from %s", err)
				errch <- err
				return
			}
			img, err := domain.Thumbnail(src, size)
			if err != nil {
				log.Errorf("Failed to create thumbnail: %s", err)
				errch <- err
				return
			}
			ch <- img
			in, out := io.Pipe()
			defer out.Close()
			go lib.thumbFormat.Encode(img, out)
			if err = lib.addPhotoFile(wrap(in), lib.thumbdir, id, size.Name+".jpg"); err != nil {
				log.Errorf("Error while creating thumbnail for %s: %s", id, err)
			}
		}()
		select {
		case img := <-ch:
			return img, nil
		case err := <-errch:
			return nil, err
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return lib.thumbFormat.Decode(f)

}

func canonicalizeFilename(photo domain.Photo) (dir, filename, id string) {
	dir = photo.DateTaken().Format("2006/01/02")
	filename = fmt.Sprintf("%s.%s", photo.ID(), photo.Format().Id)
	h := mmh3.New128()
	h.Write([]byte(photo.DateTaken().Format(time.RFC3339)))
	h.Write([]byte(strings.ToLower(filename)))
	id = fmt.Sprintf("%x", h.Sum(nil))
	return
}
