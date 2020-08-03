package views

import (
	"fmt"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/domain/gps"
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
	ID        string           `json:"id"`
	Links     Links            `json:"links"`
	Name      string           `json:"name"`
	DateTaken time.Time        `json:"dateTaken,omitempty"`
	Location  *gps.Coordinates `json:"location,omitempty"`
}

type LinkProvider struct {
	patterns map[string]string
}

func (p LinkProvider) LinksFor(object domain.Identifiable) Links {
	links := make(Links)
	id := object.ID()
	for name, pattern := range p.patterns {
		links[name] = fmt.Sprintf(pattern, id)
	}
	return links
}

func PhotoFrom(p domain.Photo) Photo {
	return Photo{
		ID:        p.ID(),
		Links:     PhotoLinksFor(p),
		Name:      p.Name(),
		DateTaken: p.DateTaken(),
		Location:  p.Location(),
	}
}

var photoLinks = LinkProvider{
	patterns: map[string]string{
		"self":  "/photos/%s",
		"view":  "/photos/%s/view",
		"thumb": "/photos/%s/thumb",
	},
}

func PhotoLinksFor(p domain.Photo) Links {
	return photoLinks.LinksFor(p)
}
