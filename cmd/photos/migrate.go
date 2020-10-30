package main

import (
	"context"

	"bitbucket.org/kleinnic74/photos/index"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/tasks"
)

type migrateTask struct {
	coordinator *index.MigrationCoordinator
}

func RegisterMigrationTask(repo *tasks.TaskRepository, coordinator *index.MigrationCoordinator) {
	repo.RegisterWithProperties("migrateTask", func() tasks.Task {
		return newMigrateTask(coordinator)
	}, tasks.TaskProperties{
		RunOnStart:   true,
		UserRunnable: false,
	})
}

func newMigrateTask(coordinator *index.MigrationCoordinator) tasks.Task {
	return migrateTask{coordinator: coordinator}
}

func (t migrateTask) Describe() string {
	return "Upgrade data"
}

func (t migrateTask) Execute(ctx context.Context, _ tasks.TaskExecutor, _ library.PhotoLibrary) error {
	return t.coordinator.Migrate(ctx)
}
