package tasks

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
)

func RegisterTasks(repo *TaskRepository) {
	repo.Register("pause", createPauseTask)
	repo.Register("generatePauseTasks", createPauseTasks)
	repo.RegisterWithProperties("initialTaskGeneration", func() Task {
		return generatePauseTasks{Count: 20}
	}, TaskProperties{
		RunOnStart:   false,
		UserRunnable: true,
	})
}

type pauseTask struct {
	Duration time.Duration `json:"duration"`
}

func createPauseTask() Task {
	return pauseTask{Duration: time.Second * 60}
}

func (t pauseTask) Describe() string {
	return fmt.Sprintf("Pausing for %s", t.Duration)
}

func (t pauseTask) Execute(ctx context.Context, executor TaskExecutor, lib library.PhotoLibrary) error {
	time.Sleep(t.Duration)
	return nil
}

type generatePauseTasks struct {
	Count int `json:"count"`
}

func createPauseTasks() Task {
	return generatePauseTasks{}
}

func (t generatePauseTasks) Describe() string {
	return fmt.Sprintf("Generating %d pause tasks", t.Count)
}

func (t generatePauseTasks) Execute(ctx context.Context, executor TaskExecutor, lib library.PhotoLibrary) error {
	for i := 0; i < t.Count; i++ {
		executor.Submit(ctx, createPauseTask())
	}
	return nil
}
