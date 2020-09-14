package tasks

import (
	"reflect"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TaskRepository struct {
	taskTypes map[string]TaskDefinition
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		taskTypes: make(map[string]TaskDefinition),
	}
}

func (r *TaskRepository) Register(name string, init TaskInitFunc) {
	r.RegisterWithProperties(name, init, TaskProperties{})
}

func (r *TaskRepository) RegisterWithProperties(name string, init TaskInitFunc, properties TaskProperties) {
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
	r.taskTypes[name] = TaskDefinition{
		Name:           name,
		TaskProperties: properties,
		init:           init,
		Parameters:     parameters,
	}
}

func (r *TaskRepository) DefinedTasks() (tasks []TaskDefinition) {
	for _, d := range r.taskTypes {
		tasks = append(tasks, d)
	}
	return
}

type TaskFilter func(t TaskDefinition) bool

func (r *TaskRepository) DefinedTasksWithFilter(filter func(t TaskDefinition) bool) (tasks []TaskDefinition) {
	for _, d := range r.taskTypes {
		if filter(d) {
			tasks = append(tasks, d)
		}
	}
	return
}

func (r *TaskRepository) CreateTask(taskType string) (Task, error) {
	def, found := r.taskTypes[taskType]
	if !found {
		return nil, UndefinedTaskType(taskType)
	}
	return def.init(), nil
}
