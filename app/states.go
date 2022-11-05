package app

import (
	"context"

	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/tasks"
	"go.uber.org/zap"
)

func launchStartupTasks(ctx context.Context, tasksRepo *tasks.TaskRepository, executor tasks.TaskExecutor) {
	for _, t := range tasksRepo.DefinedTasks() {
		if t.RunOnStart {
			logging.From(ctx).Debug("Launching startup task", zap.String("task", t.Name))
			task, err := tasksRepo.CreateTask(t.Name)
			if err != nil {
				logging.From(ctx).Warn("StartupTasks", zap.Error(err))
				continue
			}
			executor.Submit(ctx, task)
		}
	}
}
