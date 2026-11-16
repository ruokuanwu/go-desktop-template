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

type TodoStore struct {
	repo        TodoRepository
	ctx         context.Context
	cancel      context.CancelFunc
	workers     sync.WaitGroup
	mu          sync.RWMutex
	state       TodoState
	subscribers map[uint64]func(TodoState)
	nextID      uint64
}

func NewTodoStore(repo TodoRepository, items []model.Todo) *TodoStore {
	ctx, cancel := context.WithCancel(context.Background())
	return &TodoStore{
		repo: repo, ctx: ctx, cancel: cancel,
		state: TodoState{Items: slices.Clone(items)}, subscribers: make(map[uint64]func(TodoState)),
	}
}

// State returns a snapshot, including an independent copy of the items.
func (s *TodoStore) State() TodoState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state := s.state
	state.Items = slices.Clone(state.Items)
	return state
}

// Subscribe delivers an independent state snapshot for each change, without an initial notification.
// Callbacks run synchronously on the notifying goroutine, outside the lock;
// they must be concurrency-safe and nonblocking. Concurrent notifications may
// arrive out of order. Unsubscribe is idempotent but does not stop callbacks
// already captured for notification.
func (s *TodoStore) Subscribe(fn func(TodoState)) func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ctx.Err() != nil {
		return func() {}
	}
	s.nextID++
	id := s.nextID
	s.subscribers[id] = fn
	return func() {
		s.mu.Lock()
		delete(s.subscribers, id)
		s.mu.Unlock()
	}
}

// Close stops accepting commands and cancels workers. In-flight results are
// discarded. Notifications already captured may still run.
func (s *TodoStore) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancel()
	clear(s.subscribers)
}

// Wait waits for workers, notifications and success callbacks. Call it after
// Close, off the UI thread and outside subscription or success callbacks.
func (s *TodoStore) Wait() { s.workers.Wait() }

// work persists changes and updates a private snapshot without holding mu.
// The snapshot is committed only if work succeeds and the store is still open.
func (s *TodoStore) mutate(work func(context.Context, *TodoState) error, success func()) {
	s.mu.Lock()
	if s.ctx.Err() != nil {
		s.mu.Unlock()
		return
	}
	if s.state.Busy {
		s.state.Message = "正在保存任务，请稍后重试。"
		notify := s.notificationLocked()
		s.mu.Unlock()
		notify()
		return
	}
	s.state.Busy, s.state.Message = true, ""
	state := s.state
	state.Items = slices.Clone(state.Items)
	s.workers.Add(1)
	notify := s.notificationLocked()
	s.mu.Unlock()
	notify()
	go func() {
		defer s.workers.Done()
		err := work(s.ctx, &state)
		s.mu.Lock()
		if s.ctx.Err() != nil {
			s.mu.Unlock()
			return
		}
		s.state.Busy = false
		if err != nil {
			s.state.Message = err.Error()
		} else {
			s.state.Message = ""
			s.state.Items = state.Items
		}
		notify := s.notificationLocked()
		s.mu.Unlock()
		if err == nil && success != nil {
			success()
		}
		notify()
	}()
}

// notificationLocked captures the state and subscribers while mu is held.
func (s *TodoStore) notificationLocked() func() {
	state := s.state
	state.Items = slices.Clone(state.Items)
	callbacks := make([]func(TodoState), 0, len(s.subscribers))
	for _, fn := range s.subscribers {
		callbacks = append(callbacks, fn)
	}
	return func() {
		for _, fn := range callbacks {
			snapshot := state
			snapshot.Items = slices.Clone(state.Items)
			fn(snapshot)
		}
	}
}

// Add invokes success on the worker after publication, outside the lock.
// UI callbacks must schedule their changes on the UI thread. A callback already
// committed with the result may still run after Close.
func (s *TodoStore) Add(title string, success func()) {
	s.mutate(func(ctx context.Context, state *TodoState) error {
		title = strings.TrimSpace(title)
		length := utf8.RuneCountInString(title)
		if !utf8.ValidString(title) || length < 1 || length > 200 {
			return errors.New("todo title must contain 1–200 Unicode characters")
		}
		created, err := s.repo.Create(ctx, title)
		if err != nil {
			return err
		}
		state.Items = append(state.Items, created)
		return nil
	}, success)
}

func (s *TodoStore) SetCompleted(id int64, completed bool) {
	s.mutate(func(ctx context.Context, state *TodoState) error {
		if err := s.repo.SetCompleted(ctx, id, completed); err != nil {
			return err
		}
		for i := range state.Items {
			if state.Items[i].ID == id {
				state.Items[i].Completed = completed
				break
			}
		}
		return nil
	}, nil)
}

func (s *TodoStore) Delete(id int64) {
	s.mutate(func(ctx context.Context, state *TodoState) error {
		if err := s.repo.Delete(ctx, id); err != nil {
			return err
		}
		state.Items = slices.DeleteFunc(state.Items, func(item model.Todo) bool { return item.ID == id })
		return nil
	}, nil)
}
