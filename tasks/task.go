package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
)

type Task interface {
	Describe() string
	Execute(context.Context, TaskExecutor, library.PhotoLibrary) error
}

type DeferredNewPhotoCallback func(ctx context.Context, photo *library.Photo) (Task, bool)

type TaskInitFunc func() Task

type TaskProperties struct {
	RunOnStart   bool
	UserRunnable bool
}
type TaskDefinition struct {
	TaskProperties
	Name       string `json:"name"`
	init       TaskInitFunc
	Parameters []string `json:"parameters,omitempty"`
}

var (
	queue []Execution

	logger = logging.From(context.Background()).Named("tasks")
)

type UndefinedTaskType string

func (err UndefinedTaskType) Error() string {
	return fmt.Sprintf("Undefined task type '%s'", string(err))
}

type ExecutionStatus string

const (
	Pending   = ExecutionStatus("pending")
	Running   = ExecutionStatus("running")
	Completed = ExecutionStatus("completed")
	Error     = ExecutionStatus("error")
)

type TaskID uint64

type Execution struct {
	ID        TaskID          `json:"id"`
	Status    ExecutionStatus `json:"status"`
	Submitted time.Time       `json:"submitted,omitempty"`
	Completed time.Time       `json:"completed,omitempty"`
	Error     error           `json:"error,omitempty"`
	Title     string          `json:"title"`
	task      Task
}

type CompletionFunc func(Execution)

type TaskExecutor interface {
	Submit(context.Context, Task) (Execution, error)
	ListTasks(context.Context) []Execution
	DrainTasks(context.Context, CompletionFunc)
}

var ErrExecutorNotRunning = errors.New("TaskExecutor is not running")

// ExecutionsBySubmission allows sorting slices of Execution by ascending submission time
type ExecutionsBySubmission []Execution

func (a ExecutionsBySubmission) Len() int           { return len(a) }
func (a ExecutionsBySubmission) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ExecutionsBySubmission) Less(i, j int) bool { return a[i].Submitted.Before(a[j].Submitted) }
