package index

import (
	"context"
	"fmt"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/tasks"
	"go.uber.org/zap"
)

type Indexer struct {
	tracker  Tracker
	executor tasks.TaskExecutor

	indexers map[Name]interface{}
}

func NewIndexer(tracker Tracker, executor tasks.TaskExecutor) *Indexer {
	return &Indexer{
		tracker:  tracker,
		executor: executor,
		indexers: make(map[Name]interface{}),
	}
}

func (indexer *Indexer) Add(ctx context.Context, photo *library.Photo) error {
	logger, ctx := logging.FromWithNameAndFields(ctx, "indexer")
	indexes, err := indexer.tracker.GetMissingIndexes(photo.ID)
	if err != nil {
		logger.Error("Failed to retrieve missing indexes", zap.Error(err))
		return err
	}
	for _, index := range indexes {
		delegate, found := indexer.indexers[index]
		if !found {
			continue
		}
		switch f := delegate.(type) {
		case tasks.DeferredNewPhotoCallback:
			task, needed := f(ctx, photo)
			if needed {
				indexer.executor.Submit(ctx, task)
			} else {
				indexer.tracker.Update(index, photo.ID, nil)
			}
		case library.NewPhotoCallback:
			indexer.tracker.Update(index, photo.ID, f(ctx, photo))
		default:
			logger.Warn("Invalid indexer", zap.String("index", string(index)))
		}
	}
	return nil
}

func (indexer *Indexer) RegisterDefered(name Name, version library.Version, init tasks.DeferredNewPhotoCallback) {
	indexer.tracker.RegisterIndex(name, version)
	indexer.indexers[name] = init
}

func (indexer *Indexer) RegisterDirect(name Name, version library.Version, init library.NewPhotoCallback) {
	indexer.tracker.RegisterIndex(name, version)
	indexer.indexers[name] = init
}

func (indexer *Indexer) GetMissingIndexes(id library.PhotoID) ([]Name, error) {
	return indexer.tracker.GetMissingIndexes(id)
}

func (indexer *Indexer) indexDeferred(ctx context.Context, photo *library.Photo, name Name) {
	delegate, found := indexer.indexers[name]
	if !found {
		return
	}
	switch f := delegate.(type) {
	case tasks.DeferredNewPhotoCallback:
		if task, needed := f(ctx, photo); needed {
			indexer.executor.Submit(ctx, &wrappedTask{tracker: indexer.tracker, name: name, id: photo.ID, task: task})
		}
	case library.NewPhotoCallback:
		indexer.executor.Submit(ctx, &deferredCallback{tracker: indexer.tracker, name: name, f: f, photo: photo})
	}
	return
}

func (indexer *Indexer) GetIndexes() (indexers []Name) {
	for k := range indexer.indexers {
		indexers = append(indexers, k)
	}
	return
}

type deferredCallback struct {
	tracker Tracker
	name    Name
	f       library.NewPhotoCallback
	photo   *library.Photo
}

func (t *deferredCallback) Describe() string {
	return fmt.Sprintf("Indexing %s to %s", t.photo.ID, t.name)
}

func (t *deferredCallback) Execute(ctx context.Context, executor tasks.TaskExecutor, lib library.PhotoLibrary) error {
	return t.tracker.Update(t.name, t.photo.ID, t.f(ctx, t.photo))
}

type wrappedTask struct {
	tracker Tracker
	name    Name
	id      library.PhotoID
	task    tasks.Task
}

func (t *wrappedTask) Describe() string {
	return t.task.Describe()
}

func (t *wrappedTask) Execute(ctx context.Context, executor tasks.TaskExecutor, lib library.PhotoLibrary) error {
	return t.tracker.Update(t.name, t.id, t.task.Execute(ctx, executor, lib))
}

type findUnindexedTask struct {
	indexer      *Indexer
	staleIndexes []Name
}

func (indexer *Indexer) RegisterTasks(repo *tasks.TaskRepository) {
	// repo.RegisterWithProperties("findUnindexed", indexer.newFindUnindexedTask, tasks.TaskProperties{
	// 	RunOnStart:   true,
	// 	UserRunnable: false,
	// })
}

func (indexer *Indexer) NewFindUnindexedTask(staleIndexes []Name) tasks.Task {
	return findUnindexedTask{indexer, staleIndexes}
}

func (t findUnindexedTask) Describe() string {
	return "Looking for unindexed photos"
}

func (t findUnindexedTask) Execute(ctx context.Context, executor tasks.TaskExecutor, lib library.PhotoLibrary) error {
	logger, ctx := logging.SubFrom(ctx, "findUnindexedTask")
	photos, err := lib.FindAll(ctx, consts.Ascending)
	if err != nil {
		return err
	}
	var count int
	for _, p := range photos {
		missing, err := t.indexer.GetMissingIndexes(p.ID)
		if err != nil {
			logger.Warn("Could not retrieve missing indexes", zap.Error(err))
			continue
		}
		// Also add the stale indexes
		missing = append(missing, t.staleIndexes...)
		indexes := make(map[Name]bool)
		for _, i := range missing {
			if indexes[i] {
				continue
			}
			t.indexer.indexDeferred(ctx, p, i)
			logger.Info("Indexing needed", zap.String("photo", string(p.ID)), zap.String("index", string(i)))
			indexes[i] = true
		}
		if len(missing) > 0 {
			count++
		}
	}
	logger.Info("Index scan done", zap.Int("needIndexing", count))
	return nil
}
