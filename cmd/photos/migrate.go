package main

import (
	"context"
	"fmt"

	"bitbucket.org/kleinnic74/photos/index"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/tasks"
	"go.uber.org/zap"
)

type migrateTask struct {
	indexer     *index.Indexer
	coordinator *index.MigrationCoordinator

	count int
	done  int
}

func RegisterMigrationTask(repo *tasks.TaskRepository, coordinator *index.MigrationCoordinator, indexer *index.Indexer) {
	repo.RegisterWithProperties("migrateTask", func() tasks.Task {
		return newMigrateTask(coordinator, indexer)
	}, tasks.TaskProperties{
		RunOnStart:   true,
		UserRunnable: false,
	})
}

func newMigrateTask(coordinator *index.MigrationCoordinator, indexer *index.Indexer) tasks.Task {
	return &migrateTask{coordinator: coordinator, indexer: indexer}
}

func (t migrateTask) Describe() string {
	if t.count == 0 {
		return "Migrating data"
	} else {
		return fmt.Sprintf("Migrating data (%d of %d done)", t.done, t.count)
	}
}

func (t *migrateTask) Execute(ctx context.Context, executor tasks.TaskExecutor, _ library.PhotoLibrary) error {
	logger, ctx := logging.SubFrom(ctx, "migrationTask")
	staleIndexes, err := t.coordinator.Migrate(ctx, func(i int, total int) {
		t.done = i
		t.count = total
	})
	if err != nil {
		logger.Error("Error while migrating data", zap.Error(err))
		return err
	}
	updateIndexes := t.indexer.NewFindUnindexedTask(staleIndexes)
	_, err = executor.Submit(ctx, updateIndexes)
	if err != nil {
		logger.Warn("Index update task submission failed", zap.Error(err))
	}
	return err
}
