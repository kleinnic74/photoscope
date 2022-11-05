package tasks

import (
	"context"
	"fmt"
	"sync"
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

	photos library.PhotoLibrary // TODO: this should not be a field of the executor
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

func (t *serialTaskExecutor) DrainTasks(ctx context.Context, completed CompletionFunc) {
	logger := logging.From(ctx).Named("TaskExecutor")
	queue := make(map[TaskID]Execution)
	t.submitCh = make(chan taskSubmission)
	t.queryCh = make(chan executionQuery)
	taskCh := make(chan Execution)
	resCh := make(chan Execution)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			log := logger.Named(fmt.Sprintf("Worker-%d", id))
			for e := range taskCh {
				log.Info("Executing task", zap.Uint64("taskID", uint64(e.ID)))
				_, taskCtx := logging.FromWithNameAndFields(ctx, "task", zap.Uint64("taskID", uint64(e.ID)))
				e.Error = e.task.Execute(taskCtx, t, t.photos)
				if e.Error != nil {
					e.Status = Error
				} else {
					e.Status = Completed
				}
				resCh <- e
			}
			log.Info("Terminating")
		}(i)
	}
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
		wg.Wait()
	}()
	var pending []Execution
	t.running = true
	defer func() { t.running = false }()
	for {
		select {
		case s := <-t.submitCh:
			id := t.ids
			t.ids = t.ids + 1
			logger.Info("Task submitted", zap.String("task", s.task.Describe()), zap.Uint64("taskID", uint64(id)))
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
			if res.Status == Running {
				// Progress update
				queue[res.ID] = res
			} else {
				logger.Info("Task completed", zap.String("task", res.task.Describe()),
					zap.Uint64("taskID", uint64(res.ID)),
					zap.String("taskStatus", string(res.Status)),
					zap.Error(res.Error))
				completed(res)
				delete(queue, res.ID)
				if len(pending) > 0 {
					// Pick the next task from the pending queue
					e := pending[0]
					pending = pending[1:]
					e.Status = Running
					taskCh <- e
					queue[e.ID] = e
				}
			}
		case q := <-t.queryCh:
			executions := []Execution{}
			for _, v := range queue {
				v.Title = v.task.Describe()
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
