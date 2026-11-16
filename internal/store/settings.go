package store

import (
	"context"
	"sync"

	"go-desktop-template/internal/model"
)

type SettingsState struct {
	Value       model.Settings
	Busy        bool
	Message     string
	TrayMessage string
}

type SettingsService interface {
	Save(model.Settings) error
}

type SettingsStore struct {
	service SettingsService
	update  func(func())
	ctx     context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
	state   SettingsState
}

func NewSettingsStore(service SettingsService, update func(func()), initial model.Settings) *SettingsStore {
	ctx, cancel := context.WithCancel(context.Background())
	return &SettingsStore{service: service, update: update, ctx: ctx, cancel: cancel, state: SettingsState{Value: initial}}
}

func (s *SettingsStore) State() SettingsState { return s.state }

// Close stops accepting commands and cancels workers. Call it on the UI thread.
func (s *SettingsStore) Close() { s.cancel() }

// Wait waits for workers, not queued UI callbacks. Call it after Close, off the UI thread.
func (s *SettingsStore) Wait() { s.workers.Wait() }

func (s *SettingsStore) SetTrayError(err error) {
	s.state.TrayMessage = "系统托盘不可用，关闭窗口将退出程序：" + err.Error()
}

func (s *SettingsStore) SetTheme(theme string) {
	value := s.state.Value
	value.Theme = theme
	s.saveValue(value)
}

func (s *SettingsStore) SetCloseToTray(enabled bool) {
	value := s.state.Value
	value.CloseToTray = enabled
	s.saveValue(value)
}

func (s *SettingsStore) saveValue(value model.Settings) {
	if s.state.Busy || s.ctx.Err() != nil {
		return
	}
	s.state.Busy, s.state.Message = true, ""
	s.workers.Add(1)
	go func() {
		defer s.workers.Done()
		err := s.ctx.Err()
		if err == nil {
			err = s.service.Save(value)
		}
		s.update(func() {
			if s.ctx.Err() != nil {
				return
			}
			s.state.Busy = false
			if err != nil {
				s.state.Message = err.Error()
				return
			}
			s.state.Value = value
		})
	}()
}
