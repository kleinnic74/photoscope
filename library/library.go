package library

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/reusee/mmh3"
)

const (
	defaultDirMode = 0755
	DbName         = "photos.db"
)

var (
	PhotoAlreadyExists error
)

type PhotoLibrary interface {
	Add(photo domain.Photo) error
	FindAll() []domain.Photo
	Find(start, end time.Time) []domain.Photo
}

type Store interface {
	Exists(dateTaken time.Time, id string) bool
	Add(*LibraryPhoto) error
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
}

type Location struct {
	long float64 `json:"long"`
	lat  float64 `json:"lat"`
}

type LibraryPhoto struct {
	basedir   string
	path      string
	id        string
	format    *domain.Format
	dateTaken time.Time
	location  gps.Coordinates
}

func init() {
	PhotoAlreadyExists = fmt.Errorf("Photo already exists in library")
}

func (p *LibraryPhoto) Id() string {
	return p.id
}

func (p *LibraryPhoto) Format() *domain.Format {
	return p.format
}

func (p *LibraryPhoto) DateTaken() time.Time {
	return p.dateTaken
}

func (p *LibraryPhoto) Location() *gps.Coordinates {
	return &p.location
}

func (p *LibraryPhoto) Content() (io.ReadCloser, error) {
	return os.Open(filepath.Join(p.basedir, p.path))
}

func (p *LibraryPhoto) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Path      string           `json:"path"`
		Id        string           `json:"id"`
		Format    string           `json:"format"`
		DateTaken *time.Time       `json:"date"`
		Location  *gps.Coordinates `json:"gps"`
	}{
		Path:      p.path,
		Id:        p.id,
		Format:    p.format.Id,
		DateTaken: &p.dateTaken,
		Location:  &p.location,
	})
}

func (p *LibraryPhoto) UnmarshalJSON(buf []byte) error {
	data := struct {
		Path      string          `json:"path"`
		Id        string          `json:"id"`
		Format    string          `json:"format"`
		DateTaken time.Time       `json:"date"`
		Location  gps.Coordinates `json:"gps"`
	}{}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		return err
	}
	p.id = data.Id
	p.format = domain.MustFormatForExt(data.Format)
	p.path = data.Path
	p.dateTaken = data.DateTaken
	p.location = data.Location
	return nil
}

func NewBasicPhotoLibrary(basedir string, store StoreBuilder) (*BasicPhotoLibrary, error) {
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
	db, err := store(absdir, DbName)
	if err != nil {
		return nil, err
	}
	var lib = &BasicPhotoLibrary{
		basedir:  absdir,
		photodir: filepath.Join(absdir, "photos"),
		dirMode:  defaultDirMode,
		db:       db,
	}
	return lib, nil
}

func (lib *BasicPhotoLibrary) Add(photo domain.Photo) error {
	targetDir, name, id := canonicalizeFilename(photo)
	if lib.db.Exists(photo.DateTaken().UTC(), id) {
		return PhotoAlreadyExists
	}
	if err := lib.createDirectory(targetDir); err != nil {
		return err
	}
	if err := lib.addPhotoFile(photo, targetDir, name); err != nil {
		return err
	}
	p := &LibraryPhoto{
		basedir:   lib.basedir,
		path:      filepath.Join(targetDir, name),
		id:        id,
		dateTaken: photo.DateTaken().UTC(),
		location:  *photo.Location(),
		format:    photo.Format(),
	}
	return lib.db.Add(p)
}

func (lib *BasicPhotoLibrary) FindAll() []domain.Photo {
	var result []domain.Photo = make([]domain.Photo, 0)
	for _, p := range lib.db.FindAll() {
		p.basedir = lib.basedir
		result = append(result, p)
	}
	return result
}

func (lib *BasicPhotoLibrary) Find(start, end time.Time) []domain.Photo {
	var result []domain.Photo = make([]domain.Photo, 0)
	for _, p := range lib.db.Find(start, end) {
		p.basedir = lib.basedir
		result = append(result, p)
	}
	return result
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

func (lib *BasicPhotoLibrary) addPhotoFile(photo domain.Photo, targetDir, targetName string) error {
	content, err := photo.Content()
	if err != nil {
		return err
	}
	defer content.Close()
	pathInLib := filepath.Join(lib.basedir, targetDir, targetName)
	out, err := os.Create(pathInLib)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, content)
	if err != nil {
		return err
	}
	return nil
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
