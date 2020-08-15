package tasks

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Task interface {
	Execute(context.Context, TaskExecutor, library.PhotoLibrary) error
}

type TaskInitFunc func() Task

type TaskDefinition struct {
	Name       string `json:"name"`
	init       TaskInitFunc
	Parameters []string `json:"parameters,omitempty"`
}

var (
	taskTypes = map[string]TaskDefinition{}
	ids       uint64
	queue     []Execution

	logger = logging.From(context.Background()).Named("tasks")
)

func Register(name string, init TaskInitFunc) {
	taskType := init()
	t := reflect.TypeOf(taskType)
	switch t.Kind() {
	case reflect.Ptr:
		t = t.Elem()
	}
	log := logger.With(zap.String("type", name), zap.String("corrID", uuid.New().String()))
	log.Info("Task type registered")
	var parameters []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag, found := field.Tag.Lookup("json")
		if !found || len(tag) == 0 {
			continue
		}
		parameter := strings.SplitN(tag, ",", 2)[0]
		parameters = append(parameters, parameter)
		log.Info("Task parameter", zap.String("param", parameter))
	}
	taskTypes[name] = TaskDefinition{Name: name, init: init, Parameters: parameters}
}

func DefinedTasks() (tasks []TaskDefinition) {
	for _, d := range taskTypes {
		tasks = append(tasks, d)
	}
	return
}

type UndefinedTaskType string

func (err UndefinedTaskType) Error() string {
	return fmt.Sprintf("Undefined task type '%s'", err)
}

func CreateTask(taskType string) (Task, error) {
	def, found := taskTypes[taskType]
	if !found {
		return nil, UndefinedTaskType(taskType)
	}
	return def.init(), nil
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
	task      Task
}

type TaskExecutor interface {
	Submit(context.Context, Task) (Execution, error)
	ListTasks(context.Context) []Execution
	DrainTasks(context.Context)
}

var ErrExecutorNotRunning = errors.New("TaskExecutor is not running")
