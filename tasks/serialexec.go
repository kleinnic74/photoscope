package tasks

import (
	"context"
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

type taskSubmission struct {
	task      Task
	exec      chan<- Execution
	submitted time.Time
}

type executionQuery chan<- []Execution

type serialTaskExecutor struct {
	ids      TaskID
	submitCh chan taskSubmission
	queryCh  chan executionQuery
	running  bool

	photos library.PhotoLibrary
}

func NewSerialTaskExecutor(photos library.PhotoLibrary) TaskExecutor {
	return &serialTaskExecutor{
		photos: photos,
	}
}

func (t *serialTaskExecutor) Submit(ctx context.Context, task Task) (Execution, error) {
	if !t.running {
		return Execution{}, ErrExecutorNotRunning
	}
	ch := make(chan Execution)

	s := taskSubmission{task: task, exec: ch, submitted: time.Now()}
	t.submitCh <- s
	return <-ch, nil
}

func (t *serialTaskExecutor) DrainTasks(ctx context.Context) {
	logger := logging.From(ctx).Named("TaskExecutor")
	queue := make(map[TaskID]Execution)
	t.submitCh = make(chan taskSubmission)
	t.queryCh = make(chan executionQuery)
	taskCh := make(chan Execution)
	resCh := make(chan Execution)
	go func() {
		log := logger.Named("Worker")
		for e := range taskCh {
			log.Info("Executing task", zap.Uint64("taskID", uint64(e.ID)))
			e.Error = e.task.Execute(ctx, t, t.photos)
			if e.Error != nil {
				e.Status = Error
			} else {
				e.Status = Completed
			}
			resCh <- e
		}
		log.Info("Terminating")
	}()
	defer func() {
		close(taskCh)
		for s := range t.submitCh {
			close(s.exec)
		}
		close(t.submitCh)
		for q := range t.queryCh {
			close(q)
		}
		close(t.queryCh)
	}()
	var pending []Execution
	t.running = true
	defer func() { t.running = false }()
	for {
		select {
		case s := <-t.submitCh:
			id := t.ids
			t.ids = t.ids + 1
			logger.Info("Task submitted", zap.Any("task", s.task), zap.Uint64("taskID", uint64(id)))
			e := Execution{ID: id, Status: Pending, Submitted: s.submitted, task: s.task, Title: s.task.Describe()}
			select {
			case taskCh <- e:
				e.Status = Running
			default:
				// Cannot submit task yet
				pending = append(pending, e)
			}
			queue[id] = e
			s.exec <- e
			close(s.exec)
		case res := <-resCh:
			logger.Info("Task completed", zap.Any("task", res.task),
				zap.Uint64("taskID", uint64(res.ID)),
				zap.String("taskStatus", string(res.Status)),
				zap.Error(res.Error))
			delete(queue, res.ID)
			if len(pending) > 0 {
				e := pending[0]
				pending = pending[1:]
				e.Status = Running
				taskCh <- e
				queue[e.ID] = e
			}
		case q := <-t.queryCh:
			executions := []Execution{}
			for _, v := range queue {
				executions = append(executions, v)
			}
			q <- executions
			close(q)
		case <-ctx.Done():
			logger.Info("Task executor interrupted")
			return
		}
	}
}

func (t *serialTaskExecutor) ListTasks(ctx context.Context) []Execution {
	resCh := make(chan []Execution)
	t.queryCh <- resCh

	return <-resCh
}
