package tasks

import (
	"context"
	"os"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

type importFileTask struct {
	Path   string `json:"path,omitempty"`
	DryRun bool   `json:"dryrun"`
	Delete bool   `json:"delete,omitempty"`
}

func init() {
	Register("importFile", NewImportFileTask)
}

func NewImportFileTask() Task {
	return &importFileTask{}
}

func NewImportFileTaskWithParams(dryrun bool, path string, deleteAfterImport bool) Task {
	return &importFileTask{
		Path:   path,
		DryRun: dryrun,
		Delete: deleteAfterImport,
	}
}

func (t importFileTask) Execute(ctx context.Context, tasks TaskExecutor, lib library.PhotoLibrary) error {
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
	if err := lib.Add(ctx, img); err != nil {
		return err
	}

	// TODO: Create thumb

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
