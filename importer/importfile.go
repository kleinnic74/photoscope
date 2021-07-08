package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/tasks"
	"go.uber.org/zap"
)

type importFileTask struct {
	Path   string `json:"path,omitempty"`
	DryRun bool   `json:"dryrun"`
	Delete bool   `json:"delete,omitempty"`
}

func NewImportFileTask() tasks.Task {
	return &importFileTask{}
}

func NewImportFileTaskWithParams(dryrun bool, path string, deleteAfterImport bool) tasks.Task {
	return &importFileTask{
		Path:   path,
		DryRun: dryrun,
		Delete: deleteAfterImport,
	}
}

func (t importFileTask) Describe() string {
	return fmt.Sprintf("Importing file %s", t.Path)
}

func (t importFileTask) Execute(ctx context.Context, tasks tasks.TaskExecutor, lib library.PhotoLibrary) error {
	log := logging.From(ctx).Named("import")
	img, err := domain.NewPhoto(t.Path)
	if err != nil {
		log.Debug("Skipping", zap.String("file", t.Path), zap.NamedError("cause", err))
		return nil
	}
	log.Info("Found image", zap.String("file", t.Path))
	if t.DryRun {
		return nil
	}
	content, err := img.Content()
	if err != nil {
		return err
	}
	defer content.Close()

	meta := library.PhotoMeta{
		Name:        t.Path,
		Format:      img.Format(),
		Orientation: img.Orientation(),
		DateTaken:   img.DateTaken(),
		Location:    img.Location(),
	}
	if err := lib.Add(ctx, meta, content); err != nil {
		return err
	}

	if t.Delete {
		err = os.Remove(t.Path)
		if err != nil {
			log.Warn("Delete failed", zap.String("file", t.Path), zap.Error(err))
			return err
		}
		log.Info("Deleted file", zap.String("file", t.Path))
	}
	return nil
}

type importURLTask struct {
	MetaURL string
	URL     string
}

func NewImportURLTask(metaURL, contentURL string) tasks.Task {
	return &importURLTask{metaURL, contentURL}
}

func (t *importURLTask) Describe() string {
	return fmt.Sprintf("Importing %s", t.URL)
}

func (t *importURLTask) Execute(ctx context.Context, _ tasks.TaskExecutor, lib library.PhotoLibrary) error {
	photo, err := loadMeta(t.MetaURL)
	if err != nil {
		return err
	}
	resp, err := http.Get(t.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return lib.Add(ctx, photo, resp.Body)
}

func loadMeta(url string) (meta library.PhotoMeta, err error) {
	var resp *http.Response
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&meta)
	return
}
