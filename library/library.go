package library

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
)

const (
	defaultDirMode = 0755
)

type PhotoLibrary interface {
	Add(photo *domain.Photo) error
	FindAll() []domain.Photo
}

type BasicPhotoLibrary struct {
	basedir string
	dirMode os.FileMode
}

type Location struct {
	long float64 `json:"long"`
	lat  float64 `json:"lat"`
}

type libraryPhoto struct {
	path      string
	dateTaken time.Time
	location  *domain.Coordinates
}

func NewBasicPhotoLibrary(basedir string) (*BasicPhotoLibrary, error) {
	absdir, err := filepath.Abs(basedir)
	if err != nil {
		return nil, err
	}
	if info, err := os.Stat(absdir); err != nil {
		err = os.MkdirAll(absdir, 0755)
		if err != nil {
			return nil, err
		}
	} else if !info.IsDir() {
		err = fmt.Errorf("%s exists but is not a directory", absdir)
		return nil, err
	}
	var lib = &BasicPhotoLibrary{
		basedir: absdir,
		dirMode: defaultDirMode,
	}
	return lib, nil
}

func (lib *BasicPhotoLibrary) Add(photo *domain.Photo) error {
	targetDir, name, id := canonicalizeFilename(photo)
	if err := lib.createDirectory(targetDir); err != nil {
		return err
	}
	if err := lib.addPhotoFile(photo.Path, targetDir, name); err != nil {
		return err
	}
	lb := &libraryPhoto{
		path: filepath.Join(targetDir, name),
		dateTaken: photo.DateTaken.UTC(),
		location: photo.Location
	}
	return nil
}

func (lib *BasicPhotoLibrary) createDirectory(dir string) error {
	fullpath := filepath.Join(lib.basedir, dir)
	if info, err := os.Stat(fullpath); err != nil {
		return os.MkdirAll(fullpath, lib.dirMode)
	} else if !info.IsDir() {
		return fmt.Errorf("Error: %s exists but is not a directory", fullpath)
	} else {
		return nil
	}
}

func (lib *BasicPhotoLibrary) addPhotoFile(path, targetDir, targetName string) error {
	pathInLib := filepath.Join(lib.basedir, targetDir, targetName)
	if err := os.Link(path, pathInLib); err != nil {
		return err
	}
	return nil
}

func idOfPhoto(photo *domain.Photo) string {
	h := sha1.New()
	h.Write([]byte(photo.DateTaken.UTC().Format(time.RFC3339)))
	h.Write([]byte("-"))
	h.Write([]byte(strings.ToLower(photo.Filename)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func canonicalizeFilename(photo *domain.Photo) (dir, filename, id string) {
	dir = photo.DateTaken.Format("2006/01/02")
	id = idOfPhoto(photo)
	filename = fmt.Sprintf("%s.%s", id, photo.Format)
	return
}
