package tasks

import (
	"context"
	"os"
	"path/filepath"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

type importDirTask struct {
	Importdir string `json:"importdir,omitempty"`
	DryRun    bool   `json:"dryrun"`
}

var (
	skipped = map[string]struct{}{
		"@eadir": {},
	}
)

func init() {
	Register("importDir", NewImportDirTask)
}

func NewImportDirTask() Task {
	return &importDirTask{}
}

func NewImportTaskWithParams(dryrun bool, dir string) Task {
	return &importDirTask{
		Importdir: dir,
		DryRun:    dryrun,
	}
}

func (t importDirTask) Execute(ctx context.Context, tasks TaskExecutor, lib library.PhotoLibrary) error {
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
			return t.importImage(ctx, path, tasks)
		})
	} else {
		return t.importImage(ctx, t.Importdir, tasks)
	}
}

func (t importDirTask) importImage(ctx context.Context, path string, tasks TaskExecutor) error {
	task := NewImportFileTaskWithParams(false, path, false)
	_, err := tasks.Submit(ctx, task)
	return err
}
