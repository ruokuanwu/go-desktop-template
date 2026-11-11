package task

import (
	"context"
	"sync"
)

// Task publishes worker results through the supplied update function.
// Run and Close must be called on the same thread that executes published callbacks.
type Task struct {
	ctx     context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
	update  func(func())
}

func New(update func(func())) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	return &Task{ctx: ctx, cancel: cancel, update: update}
}

func (t *Task) Run(work func(context.Context) error, publish func(error)) {
	t.workers.Add(1)
	go func() {
		defer t.workers.Done()
		err := work(t.ctx)
		t.update(func() {
			if t.ctx.Err() == nil {
				publish(err)
			}
		})
	}()
}

func (t *Task) Close() { t.cancel() }
func (t *Task) Wait()  { t.workers.Wait() }

func (t *Task) Err() error { return t.ctx.Err() }
