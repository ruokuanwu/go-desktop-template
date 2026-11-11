package store

import (
	"context"

	"go-desktop-template/internal/pkg/task"

	"go-desktop-template/internal/model"
)

type Service interface {
	List(context.Context) ([]model.Todo, error)
	Create(context.Context, string) (model.Todo, error)
	SetCompleted(context.Context, int64, bool) error
	Delete(context.Context, int64) error
}

type TodoState struct {
	Items   []model.Todo
	Busy    bool
	Message string
}

type Todo struct {
	service Service
	jobs    *task.Task
	state   TodoState
}

func NewTodo(service Service, jobs *task.Task, items []model.Todo) *Todo {
	return &Todo{service: service, jobs: jobs, state: TodoState{Items: items}}
}

// State is read on the UI thread; callers must not mutate the returned items.
func (s *Todo) State() TodoState { return s.state }

func (s *Todo) mutate(work func(context.Context) error, success func()) {
	if s.state.Busy || s.jobs.Err() != nil {
		return
	}
	s.state.Busy, s.state.Message = true, ""
	var items []model.Todo
	s.jobs.Run(func(ctx context.Context) error {
		if err := work(ctx); err != nil {
			return err
		}
		var err error
		items, err = s.service.List(ctx)
		return err
	}, func(err error) {
		s.state.Busy = false
		if err != nil {
			s.state.Message = err.Error()
			return
		}
		s.state.Items = items
		if success != nil {
			success()
		}
	})
}

func (s *Todo) Add(title string, success func()) {
	s.mutate(func(ctx context.Context) error {
		_, err := s.service.Create(ctx, title)
		return err
	}, success)
}

func (s *Todo) SetCompleted(id int64, completed bool) {
	s.mutate(func(ctx context.Context) error { return s.service.SetCompleted(ctx, id, completed) }, nil)
}

func (s *Todo) Delete(id int64) {
	s.mutate(func(ctx context.Context) error { return s.service.Delete(ctx, id) }, nil)
}
