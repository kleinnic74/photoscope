package tasks

import (
	"context"
	"time"
)

type dummyexec struct {
	nextID     TaskID
	executions []Execution
}

func NewDummyTaskExecutor() TaskExecutor {
	return &dummyexec{executions: []Execution{}}
}

func (exec *dummyexec) Submit(ctx context.Context, t Task) (Execution, error) {
	e := Execution{ID: exec.nextID, task: t, Status: Pending}
	exec.nextID++
	exec.executions = append(exec.executions, e)
	return e, nil
}

func (exec *dummyexec) ListTasks(ctx context.Context) []Execution {
	return exec.executions
}

func (exec *dummyexec) DrainTasks(ctx context.Context) {
	for i, e := range exec.executions {
		e.Status = Completed
		e.Completed = time.Now()
		exec.executions[i] = e
	}
}
