package library

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/google/uuid"
	"github.com/reusee/mmh3"
	"go.uber.org/zap"
)

const (
	defaultDirMode = 0755
)

type NewPhotoCallback func(ctx context.Context, p *Photo) error

type LibraryID string

// BasicPhotoLibrary is a library storing photos on the filesystem
type BasicPhotoLibrary struct {
	ID       LibraryID
	basedir  string
	photodir string

	dirMode os.FileMode
	db      ClosableStore

	mediaLoader Loader

	thumbdir    string
	thumber     domain.Thumber
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
func NewBasicPhotoLibrary(basedir string, store ClosableStore, thumber domain.Thumber) (*BasicPhotoLibrary, error) {
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
	tmpDir := filepath.Join(absdir, "tmp", "loader")
	if err := os.MkdirAll(tmpDir, defaultDirMode); err != nil {
		return nil, err
	}
	idFilename := filepath.Join(absdir, "ID")
	dbid, err := loadOrCreateLibraryID(idFilename)
	if err != nil {
		return nil, err
	}
	return &BasicPhotoLibrary{
		ID:          dbid,
		basedir:     absdir,
		photodir:    photosDir,
		mediaLoader: NewLoader(tmpDir),

		dirMode: defaultDirMode,
		db:      store,

		thumbdir:    thumbsDir,
		thumbFormat: domain.MustFormatForExt("jpg"),
		thumber:     thumber,
	}, nil
}

func (lib *BasicPhotoLibrary) AddCallback(callback NewPhotoCallback) {
	lib.callbacks = append(lib.callbacks, callback)
}

// Add adds a photo to this library. If the given photo already exists, then
// an error of type PhotoAlreadyExists is returned
func (lib *BasicPhotoLibrary) Add(ctx context.Context, photo PhotoMeta, content io.Reader) error {
	ctx = logging.Context(ctx, logging.From(ctx).Named("library").With(zap.String("source", photo.Name)))
	targetDir, name, id := canonicalizeFilename(photo)
	orderedID := orderedIDOf(photo.DateTaken.UTC(), id)
	media, err := lib.mediaLoader.LoadMediaObject(name, content)
	if err != nil {
		return err
	}
	defer media.Cleanup()
	if dup, exists := lib.db.Exists(media.Hash()); exists {
		return PhotoAlreadyExists(dup)
	}
	var size int64
	if err := media.ProcessContent(func(in io.Reader) error {
		s, err := lib.addPhotoFile(ctx, in, lib.photodir, targetDir, name)
		size = s
		return err
	}); err != nil {
		return err
	}
	path := filepath.Join(targetDir, name)
	p := &Photo{
		Path: path,
		ExtendedPhotoID: ExtendedPhotoID{
			ID:     id,
			SortID: orderedID,
		},

		PhotoMeta: photo,
		Size:      size,
		Hash:      media.Hash(),
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

func (lib *BasicPhotoLibrary) FindByHash(ctx context.Context, hash BinaryHash) (*Photo, bool, error) {
	id, exists := lib.db.Exists(hash)
	if !exists {
		return nil, false, nil
	}
	photo, err := lib.Get(ctx, id)
	return photo, err == nil, err
}

// FindAll returns all photos from the underlying store
func (lib *BasicPhotoLibrary) FindAll(ctx context.Context, order consts.SortOrder) ([]*Photo, error) {
	return lib.db.FindAll(order)
}

// FindAllPaged returns maximum maxCount photos from the underlying store starting
// at start index
func (lib *BasicPhotoLibrary) FindAllPaged(ctx context.Context, start, maxCount int, order consts.SortOrder) ([]*Photo, bool, error) {
	return lib.db.FindAllPaged(start, maxCount, order)
}

// Find returns all photos stored in this library that have been taken between
// the given start and end times
func (lib *BasicPhotoLibrary) Find(ctx context.Context, start, end time.Time, order consts.SortOrder) ([]*Photo, error) {
	min, max := boundaryIDs(start, end)
	return lib.db.Find(min, max, order)
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
		defer baseImage.Close()
		thumb, err := lib.thumber.CreateThumb(baseImage, photo.Format, photo.Orientation, domain.Small)
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

var unknownIDs uint32

func canonicalizeFilename(photo PhotoMeta) (dir, filename string, id PhotoID) {
	dir = photo.DateTaken.Format("2006/01/02")
	name := photo.Name
	if name == "" {
		id := atomic.AddUint32(&unknownIDs, 1)
		name = fmt.Sprintf("%x-%x", photo.DateTaken.UnixNano(), id)
	}
	if dot := strings.LastIndex(name, "."); dot != -1 {
		name = name[0:dot]
	}
	basename := fmt.Sprintf("%s.%s", name, photo.Format.ID())
	h := mmh3.New128()
	h.Write([]byte(photo.DateTaken.Format(time.RFC3339)))
	h.Write([]byte(strings.ToLower(basename)))
	id = PhotoID(fmt.Sprintf("%x", h.Sum(nil)))
	filename = fmt.Sprintf("%s.%s", id, photo.Format.ID())
	return
}

func ComputeHash(in io.Reader) (BinaryHash, error) {
	// TODO: partly duplicates LoadContent, can it be improved?
	h := mmh3.New128()
	in = io.TeeReader(in, h)
	if _, err := io.Copy(ioutil.Discard, in); err != nil {
		return "", err
	}
	return BinaryHash(base64.StdEncoding.EncodeToString(h.Sum(nil))), nil
}

func (lib *BasicPhotoLibrary) MigrateInstances(ctx context.Context, progress func(int, int)) error {
	migrations := instanceMigrations(lib.ID)
	logger, ctx := logging.SubFrom(ctx, "upgradeDB")
	photos, err := lib.FindAll(ctx, consts.Ascending)
	if err != nil {
		return err
	}
	var count int
	for i, p := range photos {
		updated, err := migrations.Apply(ctx, *p, func() (io.ReadCloser, error) {
			return lib.openPhoto(p.Path)
		})
		if err != nil {
			logger.Warn("Migration failed", zap.String("photo", string(p.ID)))
			continue
		}

		if updated.schema != p.schema {
			logger.Info("Photo migrated", zap.String("photo", string(updated.ID)),
				zap.Int("pre_schema", int(p.schema)),
				zap.Int("new_schema", int(updated.schema)),
				zap.Any("content", updated))
			lib.db.Update(&updated)
			count++
		}
		progress(i, len(photos))
	}
	if count > 0 {
		logger.Info("Fixed photos in DB", zap.Int("count", count))
	}
	return nil
}

func migratePath(ctx context.Context, p Photo, _ ReaderFunc) (Photo, error) {
	logger, ctx := logging.SubFrom(ctx, "migratePath")
	if !isPathConversionNeeded(p.Path) {
		return p, nil
	}
	oldPath := p.Path
	p.Path = convertPath(oldPath)
	logger.Info("Fixed photo path", zap.String("photo", string(p.ID)), zap.String("path", p.Path), zap.String("oldpath", oldPath))
	return p, nil
}

func migrateHash(ctx context.Context, p Photo, in ReaderFunc) (Photo, error) {
	if !p.HasHash() {
		content, err := in()
		if err != nil {
			return p, err
		}
		defer content.Close()
		h, err := ComputeHash(content)
		if err != nil {
			return p, err
		}
		p.Hash = h
	}
	return p, nil
}

func addOrientation(ctx context.Context, p Photo, in ReaderFunc) (Photo, error) {
	if p.Orientation == domain.UnknownOrientation {
		content, err := in()
		if err != nil {
			return p, err
		}
		defer content.Close()
		var meta domain.MediaMetaData
		if err := p.Format.DecodeMetaData(content, &meta); err != nil {
			return p, err
		}
		p.Orientation = meta.Orientation
	}
	return p, nil
}

func addSortID(ctx context.Context, p Photo, in ReaderFunc) (Photo, error) {
	log, ctx := logging.SubFrom(ctx, "addSortID")
	if len(p.SortID) == 0 {
		p.SortID = orderedIDOf(p.DateTaken.UTC(), p.ID)
		log.Debug("Added sortID", zap.String("photo", string(p.ID)), zap.Int("sortIDSize", len(p.SortID)))
	}
	return p, nil
}

func addStoreID(libraryID LibraryID) InstanceFunc {
	return func(ctx context.Context, p Photo, in ReaderFunc) (Photo, error) {
		p.Store = libraryID
		return p, nil
	}
}

func loadOrCreateLibraryID(idFilename string) (LibraryID, error) {
	idStr, err := os.ReadFile(idFilename)
	if os.IsNotExist(err) {
		id, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}
		idStr, err := id.MarshalText()
		if err != nil {
			return "", err
		}
		f, err := os.Create(idFilename)
		if err != nil {
			return "", err
		}
		defer f.Close()
		_, err = f.Write(idStr)
		return LibraryID(idStr), err
	}
	if err != nil {
		return "", err
	}
	return LibraryID(idStr), nil
}

func instanceMigrations(libraryID LibraryID) InstanceMigrations {
	migrations := NewInstanceMigrations()
	migrations.Register(Version(1), InstanceFunc(migratePath))
	migrations.Register(Version(1), InstanceFunc(addOrientation))
	migrations.Register(Version(3), InstanceFunc(migrateHash))
	migrations.Register(Version(6), InstanceFunc(addSortID))
	migrations.Register(Version(7), addStoreID(libraryID))
	return migrations
}
