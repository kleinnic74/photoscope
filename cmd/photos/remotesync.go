package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/swarm"
	"bitbucket.org/kleinnic74/photos/tasks"
	"go.uber.org/zap"
)

func addRemoteSync(executor tasks.TaskExecutor) swarm.PeerHandler {
	return func(ctx context.Context, p swarm.Peer) {
		logger := logging.From(ctx)
		importRemote, err := NewRemoteImport(p.URL)
		if err != nil {
			logger.Warn("Failed to attach to remote peer", zap.Error(err))
			return
		}
		executor.Submit(ctx, importRemote)
	}
}

type remoteImport struct {
	remote string
}

func NewRemoteImport(remoteURL string) (t tasks.Task, err error) {
	if _, err = url.Parse(remoteURL); err != nil {
		return
	}
	t = &remoteImport{
		remote: fmt.Sprintf("%s/photos", remoteURL),
	}
	return
}

func (i *remoteImport) Describe() string {
	return fmt.Sprintf("Syncing with remote %s", i.remote)
}

func (i *remoteImport) Execute(ctx context.Context, exec tasks.TaskExecutor, lib library.PhotoLibrary) error {
	logger, ctx := logging.SubFrom(ctx, "remoteImport")
	c := &http.Client{}
	u := fmt.Sprintf("%s?p=50", i.remote)
	var cursor string
	var count int
	for hasNext := true; hasNext; count++ {
		var p page
		logger.Info("Fetching remote page", zap.String("url", u))
		resp, err := c.Get(u)
		if err != nil {
			logger.Warn("Remote sync failed", zap.Error(err))
			return err
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			logger.Warn("Remote syncing failed", zap.Error(err))
			return err
		}
		logger.Info("Sync page", zap.Int("page", count), zap.Int("itemCount", len(p.Data)))
		for _, item := range p.Data {
			if !item.HasHash() {
				logger.Info("Importing unhashed photo", zap.String("photo", item.ID))
				continue
			}
			if _, found, _ := lib.FindByHash(ctx, item.Hash); !found {
				logger.Info("Importing remote photo", zap.String("photo", item.ID))
			}
		}
		if cursor, hasNext = p.hasNext(); hasNext {
			u = fmt.Sprintf("%s?c=%s", i.remote, cursor)
		}
	}
	return nil
}

type page struct {
	Data  []item `json:"data"`
	Links []link `json:"links"`
}

type item struct {
	ID    string             `json:"id"`
	Links map[string]string  `json:"links"`
	Hash  library.BinaryHash `json:"hash,omitempty"`
}

func (i item) HasHash() bool {
	return len(i.Hash) > 0
}

func (p page) hasNext() (string, bool) {
	for _, l := range p.Links {
		if l.Name == "next" {
			return l.Cursor, l.Cursor != ""
		}
	}
	return "", false
}

type link struct {
	Name   string `json:"name"`
	Cursor string `json:"href"`
}
