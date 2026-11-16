package store

import (
	"context"
	"sync"

	"go-desktop-template/internal/model"
)

type SettingsState struct {
	Value   model.Settings
	Busy    bool
	Message string
}

type SettingsService interface {
	Save(model.Settings) error
}

type SettingsStore struct {
	service     SettingsService
	ctx         context.Context
	cancel      context.CancelFunc
	workers     sync.WaitGroup
	mu          sync.RWMutex
	state       SettingsState
	subscribers map[uint64]func()
	nextID      uint64
}

func NewSettingsStore(service SettingsService, initial model.Settings) *SettingsStore {
	ctx, cancel := context.WithCancel(context.Background())
	return &SettingsStore{
		service: service, ctx: ctx, cancel: cancel,
		state: SettingsState{Value: initial}, subscribers: make(map[uint64]func()),
	}
}

// State returns a snapshot safe to read and modify independently.
func (s *SettingsStore) State() SettingsState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Subscribe registers a change notification without an initial notification.
// Callbacks run synchronously on the notifying goroutine, outside the lock;
// they must be concurrency-safe and nonblocking. Unsubscribe is idempotent
// but does not stop callbacks already captured for notification.
func (s *SettingsStore) Subscribe(fn func()) func() {
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

// Close stops accepting commands and cancels workers. In-flight saves may finish,
// but their results are discarded. Notifications already captured may still run.
func (s *SettingsStore) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancel()
	clear(s.subscribers)
}

// Wait waits for workers and their notifications. Call it after Close,
// off the UI thread and outside subscription callbacks.
func (s *SettingsStore) Wait() { s.workers.Wait() }

func (s *SettingsStore) SetTrayError(err error) {
	s.mu.Lock()
	if s.ctx.Err() != nil {
		s.mu.Unlock()
		return
	}
	notify := s.notificationLocked()
	s.mu.Unlock()
	notify()
}

func (s *SettingsStore) SetTheme(theme string) {
	s.saveValue(func(value *model.Settings) { value.Theme = theme })
}

func (s *SettingsStore) SetCloseToTray(enabled bool) {
	s.saveValue(func(value *model.Settings) { value.CloseToTray = enabled })
}

func (s *SettingsStore) saveValue(change func(*model.Settings)) {
	s.mu.Lock()
	if s.state.Busy || s.ctx.Err() != nil {
		s.mu.Unlock()
		return
	}
	value := s.state.Value
	change(&value)
	s.state.Busy, s.state.Message = true, ""
	s.workers.Add(1)
	notify := s.notificationLocked()
	s.mu.Unlock()
	notify()
	go func() {
		defer s.workers.Done()
		err := s.ctx.Err()
		if err == nil {
			err = s.service.Save(value)
		}
		s.mu.Lock()
		if s.ctx.Err() != nil {
			s.mu.Unlock()
			return
		}
		s.state.Busy = false
		if err != nil {
			s.state.Message = err.Error()
		} else {
			s.state.Value = value
		}
		notify := s.notificationLocked()
		s.mu.Unlock()
		notify()
	}()
}

// notificationLocked captures subscribers while mu is held.
func (s *SettingsStore) notificationLocked() func() {
	callbacks := make([]func(), 0, len(s.subscribers))
	for _, fn := range s.subscribers {
		callbacks = append(callbacks, fn)
	}
	return func() {
		for _, fn := range callbacks {
			fn()
		}
	}
}
