package main

import (
	"log"
	"os"
	"path/filepath"

	"bitbucket.org/kleinnic74/photos/domain"
)

type DirectoryImporter struct {
	basedir string
	skipped map[string]bool
}

type PhotoHandler func(domain.Photo) error

type PhotoHandlerChain struct {
	handler []PhotoHandler
}

func NewPhotoHandlerChain() *PhotoHandlerChain {
	return &PhotoHandlerChain{
		handler: make([]PhotoHandler, 0),
	}
}

func (c *PhotoHandlerChain) Then(h PhotoHandler) *PhotoHandlerChain {
	c.handler = append(c.handler, h)
	return c
}

func (c *PhotoHandlerChain) Do() PhotoHandler {
	return func(p domain.Photo) error {
		for _, h := range c.handler {
			if err := h(p); err != nil {
				log.Printf("Photo %s: %s", p.Id(), err)
			}
		}
		return nil
	}
}

func NewDirectoryImporter(dir string) (*DirectoryImporter, error) {
	absdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return &DirectoryImporter{
		basedir: absdir,
		skipped: make(map[string]bool),
	}, nil
}

func (d *DirectoryImporter) SkipDir(name string) *DirectoryImporter {
	d.skipped[name] = true
	return d
}

func (d *DirectoryImporter) Walk(handler PhotoHandler) error {
	return filepath.Walk(d.basedir, d.loadImage(handler))
}

func (d *DirectoryImporter) loadImage(handler PhotoHandler) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error while walking: %s", err)
			return err
		}
		if info.IsDir() {
			if skip, found := d.skipped[info.Name()]; found && skip {
				log.Printf("Skipping directory %s", path)
				return filepath.SkipDir
			}
		} else {
			img, err := domain.NewPhoto(path)
			if err == nil {
				handler(img)
			} else {
				log.Printf("Error: not an image: %s [%s]", path, err)
			}
		}
		return nil
	}
}
