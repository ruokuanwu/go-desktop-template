package asyncjobs

import (
	"context"
	"sync"
)

// AsyncJobs publishes worker results through the supplied update function.
// Run and Close must be called on the same thread that executes published callbacks.
type AsyncJobs struct {
	ctx     context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
	update  func(func())
}

func New(update func(func())) *AsyncJobs {
	ctx, cancel := context.WithCancel(context.Background())
	return &AsyncJobs{ctx: ctx, cancel: cancel, update: update}
}

func (j *AsyncJobs) Run(work func(context.Context) error, publish func(error)) {
	j.workers.Add(1)
	go func() {
		defer j.workers.Done()
		err := work(j.ctx)
		j.update(func() {
			if j.ctx.Err() == nil {
				publish(err)
			}
		})
	}()
}

func (j *AsyncJobs) Close() { j.cancel() }
func (j *AsyncJobs) Wait()  { j.workers.Wait() }

func (j *AsyncJobs) Err() error { return j.ctx.Err() }
