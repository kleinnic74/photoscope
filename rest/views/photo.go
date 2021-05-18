package views

import (
	"fmt"
	"time"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"bitbucket.org/kleinnic74/photos/library"
)

type Links map[string]string

func (l Links) Add(name, link string) Links {
	l[name] = link
	return l
}

func (l Links) AddAll(links Links) Links {
	for k, v := range links {
		l[k] = v
	}
	return l
}

type Photo struct {
	ID        library.PhotoID  `json:"id"`
	Links     Links            `json:"links"`
	Name      string           `json:"name"`
	Hash      string           `json:"hash,omitempty"`
	DateTaken time.Time        `json:"dateTaken,omitempty"`
	Location  *gps.Coordinates `json:"location,omitempty"`
}

type LinkProvider struct {
	patterns map[string]string
}

func (p LinkProvider) LinksFor(photo library.PhotoID) Links {
	links := make(Links)
	for name, pattern := range p.patterns {
		links[name] = fmt.Sprintf(pattern, photo)
	}
	return links
}

func PhotoFrom(p *library.Photo) Photo {
	return Photo{
		ID:        p.ID,
		Links:     PhotoLinksFor(p.ID),
		Name:      p.Name(),
		Hash:      p.Hash.String(),
		DateTaken: p.DateTaken,
		Location:  p.Location,
	}
}

var photoLinks = LinkProvider{
	patterns: map[string]string{
		"self":  "/photos/%s",
		"view":  "/photos/%s/view",
		"thumb": "/photos/%s/thumb",
	},
}

func PhotoLinksFor(p library.PhotoID) Links {
	return photoLinks.LinksFor(p)
}
