package wdav

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/webdav"
)

type closeCallback func(node *fsNode) error

type fsNode struct {
	dir      bool
	name     string
	parent   *fsNode
	ctime    time.Time
	children map[string]*fsNode
	close    closeCallback
	tmpname  string
	file     *os.File
}

func newFileNode(name string, parent *fsNode, close closeCallback) *fsNode {
	return &fsNode{
		dir:     false,
		name:    name,
		parent:  parent,
		ctime:   time.Now(),
		close:   close,
		tmpname: uuid.New().String(),
	}
}

func NewDirNode(name string, parent *fsNode) *fsNode {
	return &fsNode{
		dir:      true,
		name:     name,
		parent:   parent,
		ctime:    time.Now(),
		children: make(map[string]*fsNode),
	}
}

func (d *fsNode) Child(name string) (child *fsNode, exists bool) {
	child, exists = d.children[name]
	return
}

func (d *fsNode) Add(child *fsNode) error {
	if !d.dir {
		return os.ErrInvalid
	}
	d.children[child.name] = child
	return nil
}

// Interface os.FileInfo

func (n *fsNode) IsDir() bool {
	return n.dir
}

func (n *fsNode) ModTime() time.Time {
	return n.ctime
}

func (n *fsNode) Mode() os.FileMode {
	return os.FileMode(0644)
}

func (n *fsNode) Name() string {
	return n.name
}

func (n *fsNode) Size() int64 {
	return 0
}

func (n *fsNode) Sys() interface{} {
	return nil
}

// Interface webdav.File

var n webdav.File = &fsNode{}
var h http.File = &fsNode{}

func (n *fsNode) Close() error {
	if n.file != nil {
		if err := n.file.Close(); err != nil {
			return err
		}
	}
	if n.close != nil {
		return n.close(n)
	}
	return nil
}

func (n *fsNode) Write(p []byte) (count int, err error) {
	return n.file.Write(p)
}

func (n *fsNode) Read(p []byte) (count int, err error) {
	return n.file.Read(p)
}

func (n *fsNode) Seek(offset int64, whence int) (int64, error) {
	return n.file.Seek(offset, whence)
}

func (n *fsNode) Readdir(count int) (entries []os.FileInfo, err error) {
	if !n.dir {
		err = os.ErrInvalid
		return
	}
	for _, n := range n.children {
		entries = append(entries, n)
	}
	return
}

func (n *fsNode) Stat() (os.FileInfo, error) {
	return n, nil
}

func (n *fsNode) createTmpFile(dir string) error {
	if n.dir {
		panic(fmt.Sprintf("File %s is a directory, cannot create a temporary file for it", n.name))
	}
	path := filepath.Join(dir, n.tmpname)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	n.file = f
	return nil
}
