package store

import (
	"context"
	"slices"
	"sync"

	"go-desktop-template/internal/model"
)

type TodoService interface {
	Create(context.Context, string) (model.Todo, error)
	SetCompleted(context.Context, int64, bool) error
	Delete(context.Context, int64) error
}

type TodoState struct {
	Items   []model.Todo
	Busy    bool
	Message string
}

type TodoStore struct {
	service TodoService
	update  func(func())
	ctx     context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
	state   TodoState
}

func NewTodoStore(service TodoService, update func(func()), items []model.Todo) *TodoStore {
	ctx, cancel := context.WithCancel(context.Background())
	return &TodoStore{service: service, update: update, ctx: ctx, cancel: cancel, state: TodoState{Items: items}}
}

// State is read on the UI thread; callers must not mutate the returned items.
func (s *TodoStore) State() TodoState { return s.state }

// Close stops accepting commands and cancels workers. Call it on the UI thread.
func (s *TodoStore) Close() { s.cancel() }

// Wait waits for workers, not queued UI callbacks. Call it after Close, off the UI thread.
func (s *TodoStore) Wait() { s.workers.Wait() }

// Commands and published callbacks run on the UI thread.
func (s *TodoStore) mutate(work func(context.Context) error, success func()) {
	if s.ctx.Err() != nil {
		s.state.Message = "任务存储已关闭，无法执行操作。"
		return
	}
	if s.state.Busy {
		s.state.Message = "正在保存任务，请稍后重试。"
		return
	}
	s.state.Busy, s.state.Message = true, ""
	s.workers.Add(1)
	go func() {
		defer s.workers.Done()
		err := work(s.ctx)
		s.update(func() {
			if s.ctx.Err() != nil {
				return
			}
			s.state.Busy = false
			if err != nil {
				s.state.Message = err.Error()
				return
			}
			s.state.Message = ""
			success()
		})
	}()
}

func (s *TodoStore) Add(title string, success func()) {
	var created model.Todo
	s.mutate(func(ctx context.Context) error {
		var err error
		created, err = s.service.Create(ctx, title)
		return err
	}, func() {
		s.state.Items = append(s.state.Items, created)
		if success != nil {
			success()
		}
	})
}

func (s *TodoStore) SetCompleted(id int64, completed bool) {
	s.mutate(func(ctx context.Context) error { return s.service.SetCompleted(ctx, id, completed) }, func() {
		for i := range s.state.Items {
			if s.state.Items[i].ID == id {
				s.state.Items[i].Completed = completed
				break
			}
		}
	})
}

func (s *TodoStore) Delete(id int64) {
	s.mutate(func(ctx context.Context) error { return s.service.Delete(ctx, id) }, func() {
		s.state.Items = slices.DeleteFunc(s.state.Items, func(item model.Todo) bool { return item.ID == id })
	})
}
