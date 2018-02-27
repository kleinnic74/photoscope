package library

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
)

type PhotoLibrary interface {
	Add(photo *domain.Photo)
	FindAll() []domain.Photo
}

type BasicPhotoLibrary struct {
	basedir string
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
	}
	return lib, nil
}

func (lib *BasicPhotoLibrary) Add(photo *domain.Photo) {
	targetDir := photo.DateTaken.Format("2006/01/02")
	h := sha1.New()
	h.Write([]byte(photo.DateTaken.UTC().Format(time.RFC3339)))
	h.Write([]byte("-"))
	h.Write([]byte(strings.ToLower(photo.Filename)))
	name := fmt.Sprintf("%x.%s", h.Sum(nil))
	pathInLib := filepath.Join(lib.basedir, targetDir, name)
	dir := filepath.Dir(pathInLib)
	if info, err := os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("Error: cannot create directory %s: %s", dir, err)
			return
		}
	} else if !info.IsDir() {
		log.Printf("Error: %s is not a directory", dir)
		return
	}
	if err := os.Link(photo.Path, pathInLib); err != nil {
		log.Printf("Error: cannot link %s to %s: %s", photo.Path, pathInLib, err)
	}
}
