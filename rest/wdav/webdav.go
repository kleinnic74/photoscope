package wdav

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
	"golang.org/x/net/webdav"
)

type WebDavAdapter struct {
	lib    library.PhotoLibrary
	root   *fsNode
	tmpDir string
}

func NewWebDavAdapter(lib library.PhotoLibrary, tmpdir string) (*WebDavAdapter, error) {
	info, err := os.Stat(tmpdir)
	if err != nil {
		if err = os.MkdirAll(tmpdir, 0755); err != nil {
			return nil, err
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("Path '%s' is not a directory", tmpdir)
	}
	return &WebDavAdapter{
		lib:    lib,
		root:   NewDirNode("", nil),
		tmpDir: tmpdir,
	}, nil
}

func (dav *WebDavAdapter) Mkdir(ctx context.Context, name string, perm os.FileMode) (err error) {
	// Always succeed
	parts := splitPath(name)
	logging.From(ctx).Info("Mkdir", zap.String("name", name), zap.Strings("parts", parts))
	p := dav.root
	for _, part := range parts {
		if child, exists := p.Child(part); exists {
			p = child
			continue
		}
		child := NewDirNode(part, p)
		if err = p.Add(child); err != nil {
			return
		}
	}
	return nil
}

func (dav *WebDavAdapter) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if (flag & os.O_CREATE) != 0 {
		path := splitPath(name)
		dir, filename := splitDirFromName(path)
		logging.From(ctx).Info("Create file",
			zap.String("name", name),
			zap.String("flags", strconv.FormatInt(int64(flag), 2)),
			zap.Strings("parts", path),
		)
		parent, err := dav.findNode(dir)
		if err != nil {
			return nil, err
		}
		f := newFileNode(filename, parent, func(node *fsNode) error {
			return dav.close(ctx, node)
		})
		if err := f.createTmpFile(dav.tmpDir); err != nil {
			return nil, err
		}
		parent.Add(f)
		return f, nil
	} else {
		logging.From(ctx).Info("OpenFile", zap.String("name", name), zap.String("flags", strconv.FormatInt(int64(flag), 2)))
	}
	return nil, os.ErrNotExist
}

func (dav *WebDavAdapter) RemoveAll(ctx context.Context, name string) error {
	logging.From(ctx).Info("RemoveAll", zap.String("name", name))
	return nil
}

func (dav *WebDavAdapter) Rename(ctx context.Context, oldName, newName string) error {
	logging.From(ctx).Info("Rename", zap.String("from", oldName), zap.String("to", newName))
	return nil
}

func (dav *WebDavAdapter) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	parts := splitPath(name)
	logging.From(ctx).Info("Stat", zap.String("name", name), zap.Strings("path", parts))
	p := dav.root
	for _, name := range parts {
		child, exists := p.Child(name)
		if !exists {
			return nil, os.ErrNotExist
		}
		p = child
	}
	return p, nil
}

func (dav *WebDavAdapter) findNode(path []string) (*fsNode, error) {
	node := dav.root
	for i, p := range path {
		child, found := node.Child(p)
		if !found {
			return nil, fmt.Errorf("Path does not exist: %s", path[0:i])
		}
		node = child
	}
	return node, nil
}

func (dav *WebDavAdapter) close(ctx context.Context, node *fsNode) error {
	logging.From(ctx).Info("Closed", zap.String("name", node.name))
	return nil
}
