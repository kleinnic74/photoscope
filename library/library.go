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
	DbName         = "photos.db"
)

type PhotoLibrary interface {
	Add(photo domain.Photo) error
	Get(id string) (domain.Photo, error)
	FindAll() []domain.Photo
	Find(start, end time.Time) []domain.Photo
}

type Store interface {
	Exists(dateTaken time.Time, id string) bool
	Add(*LibraryPhoto) error
	Get(id string) (*LibraryPhoto, error)
	FindAll() []*LibraryPhoto
	Find(start, end time.Time) []*LibraryPhoto
}

type ClosableStore interface {
	Store

	Close()
}

type StoreBuilder func(string, string) (ClosableStore, error)

type BasicPhotoLibrary struct {
	basedir  string
	photodir string
	dirMode  os.FileMode
	db       ClosableStore

	thumbdir    string
	thumbFormat *domain.Format
}

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
	db, err := store(absdir, DbName)
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

func (lib *BasicPhotoLibrary) Add(photo domain.Photo) error {
	targetDir, name, id := canonicalizeFilename(photo)
	if lib.db.Exists(photo.DateTaken().UTC(), id) {
		return PhotoAlreadyExists(id)
	}
	if err := lib.addPhotoFile(photo.Content, lib.photodir, targetDir, name); err != nil {
		return err
	}
	p := &LibraryPhoto{
		lib:       lib,
		path:      filepath.Join(targetDir, name),
		id:        id,
		dateTaken: photo.DateTaken().UTC(),
		location:  *photo.Location(),
		format:    photo.Format(),
	}
	return lib.db.Add(p)
}

func (lib *BasicPhotoLibrary) Get(id string) (domain.Photo, error) {
	if p, err := lib.db.Get(id); err != nil {
		return nil, err
	} else {
		p.lib = lib
		return p, err
	}
}

func (lib *BasicPhotoLibrary) FindAll() []domain.Photo {
	var result []domain.Photo = make([]domain.Photo, 0)
	for _, p := range lib.db.FindAll() {
		p.lib = lib
		result = append(result, p)
	}
	return result
}

func (lib *BasicPhotoLibrary) Find(start, end time.Time) []domain.Photo {
	var result []domain.Photo = make([]domain.Photo, 0)
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
		log.Infof("Creating thumbnail for %s at %s...", id, path)
		src, err := photo.Image()
		if err != nil {
			log.Errorf("Could not read image from %s", err)
			return nil, err
		}
		img, err := domain.Thumbnail(src, size)
		if err != nil {
			log.Errorf("Failed to create thumbnail: %s", err)
			return nil, err
		}
		in, out := io.Pipe()
		go func() {
			defer out.Close()
			lib.thumbFormat.Encode(img, out)
		}()
		if err = lib.addPhotoFile(wrap(in), lib.thumbdir, id, size.Name+".jpg"); err != nil {
			log.Errorf("Error while creating thumbnail for %s: %s", id, err)
		}
		return img, nil
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return lib.thumbFormat.Decode(f)
	}
}

func canonicalizeFilename(photo domain.Photo) (dir, filename, id string) {
	dir = photo.DateTaken().Format("2006/01/02")
	filename = fmt.Sprintf("%s.%s", photo.Id(), photo.Format().Id)
	h := mmh3.New128()
	h.Write([]byte(photo.DateTaken().Format(time.RFC3339)))
	h.Write([]byte(strings.ToLower(filename)))
	id = fmt.Sprintf("%x", h.Sum(nil))
	return
}
