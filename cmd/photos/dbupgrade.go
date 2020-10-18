package main

import (
	"context"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/tasks"
)

type upgradeDB struct {
	db *library.BasicPhotoLibrary
}

func RegisterDBUpgradeTasks(repo *tasks.TaskRepository, lib *library.BasicPhotoLibrary) {
	repo.RegisterWithProperties("upgradeDB", func() tasks.Task {
		return newUpgradeDBTask(lib)
	}, tasks.TaskProperties{
		RunOnStart:   true,
		UserRunnable: false,
	})
}

func newUpgradeDBTask(lib *library.BasicPhotoLibrary) tasks.Task {
	return upgradeDB{db: lib}
}

func (t upgradeDB) Describe() string {
	return "Fix internal DB structures"
}

func (t upgradeDB) Execute(ctx context.Context, _ tasks.TaskExecutor, _ library.PhotoLibrary) error {
	return t.db.UpgradeDBStructures(ctx)
}
