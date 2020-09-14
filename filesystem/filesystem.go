package filesystem

import (
	"io"
	"os"
	"path/filepath"
)

const defaultDirMode = 0755

type PathOperations interface {
	Path(string) string
	PathToFile(string, string) string
}

type Facade interface {
	PathOperations
	Mkdir(string) error
	Create(string, string) (io.WriteCloser, error)
}

type defaultFilesystem string

func NewDefaultFilesystem(basedir string) (Facade, error) {
	absdir, err := filepath.Abs(basedir)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(basedir, defaultDirMode); err != nil {
		return nil, err
	}
	return defaultFilesystem(absdir), nil
}

func (fs defaultFilesystem) Mkdir(path string) error {
	target := filepath.Join(string(fs), path)
	return os.MkdirAll(target, defaultDirMode)
}

func (fs defaultFilesystem) Path(subpath string) string {
	return filepath.Join(string(fs), subpath)
}

func (fs defaultFilesystem) PathToFile(subdir, filename string) string {
	return filepath.Join(string(fs), subdir, filename)
}

func (fs defaultFilesystem) Create(subdir string, name string) (io.WriteCloser, error) {
	fullpath := filepath.Join(string(fs), subdir, name)
	return os.Create(fullpath)
}
