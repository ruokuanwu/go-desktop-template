package store

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"

	"go-desktop-template/internal/model"
)

type TodoRepository interface {
	Create(context.Context, string) (model.Todo, error)
	SetCompleted(context.Context, int64, bool) error
	Delete(context.Context, int64) error
}

type TodoState struct {
	Items   []model.Todo
	Busy    bool
	Message string
}

// TodoStore owns UI state. Commands, State and Close run on the UI thread.
// post must asynchronously dispatch to that thread and request a redraw.
type TodoStore struct {
	repo    TodoRepository
	post    func(func())
	ctx     context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
	state   TodoState
}

func NewTodoStore(repo TodoRepository, items []model.Todo, post func(func())) *TodoStore {
	ctx, cancel := context.WithCancel(context.Background())
	return &TodoStore{repo: repo, post: post, ctx: ctx, cancel: cancel, state: TodoState{Items: slices.Clone(items)}}
}

// State returns shared UI-thread state, not an independent snapshot.
func (s *TodoStore) State() *TodoState { return &s.state }

// Close rejects new commands and queued results. Call it on the UI thread.
func (s *TodoStore) Close() { s.cancel() }

// Wait waits for background I/O, not queued UI callbacks. Call after Close,
// outside the UI thread so shutdown cannot block event processing.
func (s *TodoStore) Wait() { s.workers.Wait() }

func (s *TodoStore) mutate(work func(context.Context) (func(), error), success func()) {
	if s.ctx.Err() != nil {
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
		apply, err := work(s.ctx)
		s.post(func() {
			if s.ctx.Err() != nil {
				return
			}
			s.state.Busy = false
			if err != nil {
				s.state.Message = err.Error()
				return
			}
			s.state.Message = ""
			apply()
			if success != nil {
				success()
			}
		})
	}()
}

// Add invokes success on the UI thread after publishing the persisted result.
func (s *TodoStore) Add(title string, success func()) {
	s.mutate(func(ctx context.Context) (func(), error) {
		title = strings.TrimSpace(title)
		length := utf8.RuneCountInString(title)
		if !utf8.ValidString(title) || length < 1 || length > 200 {
			return nil, errors.New("todo title must contain 1–200 Unicode characters")
		}
		created, err := s.repo.Create(ctx, title)
		if err != nil {
			return nil, err
		}
		return func() { s.state.Items = append(s.state.Items, created) }, nil
	}, success)
}

func (s *TodoStore) SetCompleted(id int64, completed bool) {
	s.mutate(func(ctx context.Context) (func(), error) {
		if err := s.repo.SetCompleted(ctx, id, completed); err != nil {
			return nil, err
		}
		return func() {
			for i := range s.state.Items {
				if s.state.Items[i].ID == id {
					s.state.Items[i].Completed = completed
					break
				}
			}
		}, nil
	}, nil)
}

func (s *TodoStore) Delete(id int64) {
	s.mutate(func(ctx context.Context) (func(), error) {
		if err := s.repo.Delete(ctx, id); err != nil {
			return nil, err
		}
		return func() {
			s.state.Items = slices.DeleteFunc(s.state.Items, func(item model.Todo) bool { return item.ID == id })
		}, nil
	}, nil)
}
