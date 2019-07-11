// XXX
// * send on closed taskQ?
package mars

import (
	"context"
	"runtime"

	"golang.org/x/sync/errgroup"
)

const (
	queueSizePerWorker = 1 << 13
)

// Primitive Tasks
type Task interface {
	Run(M *MARS) error
}

type TaskFn func(ctx context.Context) error

func (fn TaskFn) Run(M *MARS) error {
	return fn(M.ctx)
}

type SchedulerTaskFn TaskFn

func (fn SchedulerTaskFn) Run(M *MARS) error {
	M.group.Go(func() error {
		return fn(M.ctx)
	})
	return nil
}

// Config
type Config struct {
	Concurrency int
	QueueSize   int
}

func (config_ *Config) Normalize() (config Config) {
	config = *config_
	if config.Concurrency < 1 {
		config.Concurrency = runtime.GOMAXPROCS(0)
	}
	if config.QueueSize < 1 {
		config.QueueSize = queueSizePerWorker * config.Concurrency
	}
	return
}

// MARS
type MARS struct {
	ctx   context.Context
	group *errgroup.Group
	taskQ chan Task
}

func New(ctx context.Context, config Config) (M *MARS) {
	config = config.Normalize()
	g, ctx := errgroup.WithContext(ctx)
	M = &MARS{
		ctx:   ctx,
		group: g,
		taskQ: make(chan Task, config.QueueSize),
	}
	for i := 0; i < config.Concurrency; i++ {
		_ = SchedulerTaskFn(M.Worker).Run(M)
	}
	return
}

func (M *MARS) Context() context.Context {
	return M.ctx
}

func (M *MARS) Worker(ctx context.Context) error {
	for {
		select {
		case task, more := <-M.taskQ:
			if !more {
				return nil
			}
			return task.Run(M)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (M *MARS) Submit(task Task) error {
	ctx := M.ctx
	select {
	case M.taskQ <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// XXX inline execution to prevent deadlock due to back pressure,
		//     but can cause stack overflow => check nesting & trampoline
		return task.Run(M)
	}
}

func (M *MARS) SubmitFn(fn func(ctx context.Context) error) error {
	return M.Submit(TaskFn(fn))
}

func (M *MARS) Shutdown() {
	close(M.taskQ)
}

func (M *MARS) Wait() error {
	return M.group.Wait()
}
