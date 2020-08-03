package tasks

import (
	"context"
	"os"
	"path/filepath"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

type importTask struct {
	Importdir string `json:"importdir,omitempty"`
	DryRun    bool   `json:"dryrun"`
}

var (
	skipped = map[string]struct{}{
		"@eadir": {},
	}
)

func init() {
	Register("import", NewImportTask)
}

func NewImportTask() Task {
	return &importTask{}
}

func NewImportTaskWithParams(dryrun bool, dir string) Task {
	return &importTask{
		Importdir: dir,
		DryRun:    dryrun,
	}
}

func (t importTask) Execute(ctx context.Context, lib library.PhotoLibrary) error {
	ctx, logger := logging.SubFrom(ctx, "importTask")
	logger.Info("Importing photos", zap.String("dir", t.Importdir))
	var count uint
	defer func() {
		logger.Info("Import finished", zap.Uint("count", count))
	}()
	stat, err := os.Stat(t.Importdir)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return filepath.Walk(t.Importdir, func(path string, info os.FileInfo, err error) error {
			logger.Debug("Visiting file", zap.String("path", path),
				zap.String("name", info.Name()))
			if _, found := skipped[info.Name()]; found && info.IsDir() {
				logger.Debug("SKipping dir", zap.String("dir", path))
				return filepath.SkipDir
			}
			if info.IsDir() {
				logger.Debug("Entering dir", zap.String("dir", path))
				return nil
			}
			return t.importImage(ctx, path, lib)
		})
	} else {
		return t.importImage(ctx, t.Importdir, lib)
	}
}

func (t importTask) importImage(ctx context.Context, path string, lib library.PhotoLibrary) error {
	log := logging.From(ctx)
	img, err := domain.NewPhoto(path)
	if err != nil {
		log.Debug("Skipping", zap.String("file", path), zap.NamedError("cause", err))
		return nil
	}
	log.Info("Found image", zap.String("file", path))
	if t.DryRun {
		return nil
	}
	if err := lib.Add(ctx, img); err != nil {
		return err
	}
	// Create thumb
	return nil
}
