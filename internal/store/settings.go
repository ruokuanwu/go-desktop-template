package store

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-desktop-template/internal/model"
)

type SettingsState struct {
	Value   model.Settings
	Busy    bool
	Message string
}

type SettingsRepository interface {
	Save(model.Settings) error
}

// SettingsStore owns UI state. Commands, State and Close run on the UI thread.
// post must asynchronously dispatch to that thread and request a redraw.
type SettingsStore struct {
	repo    SettingsRepository
	post    func(func())
	ctx     context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
	state   SettingsState
}

func NewSettingsStore(repo SettingsRepository, initial model.Settings, post func(func())) (*SettingsStore, error) {
	if err := validateSettings(initial); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &SettingsStore{repo: repo, post: post, ctx: ctx, cancel: cancel, state: SettingsState{Value: initial}}, nil
}

// State returns shared UI-thread state, not an independent snapshot.
func (s *SettingsStore) State() *SettingsState { return &s.state }

// Close rejects new commands and queued results. Call it on the UI thread.
func (s *SettingsStore) Close() { s.cancel() }

// Wait waits for background I/O, not queued UI callbacks. Call after Close,
// outside the UI thread so shutdown cannot block event processing.
func (s *SettingsStore) Wait() { s.workers.Wait() }

func (s *SettingsStore) SetLanguage(language string) {
	s.saveValue(func(value *model.Settings) { value.Language = language })
}

func (s *SettingsStore) SetTheme(theme string) {
	s.saveValue(func(value *model.Settings) { value.Theme = theme })
}

func (s *SettingsStore) SetCloseToTray(enabled bool) {
	s.saveValue(func(value *model.Settings) { value.CloseToTray = enabled })
}

func (s *SettingsStore) saveValue(change func(*model.Settings)) {
	if s.state.Busy || s.ctx.Err() != nil {
		return
	}
	value := s.state.Value
	change(&value)
	s.state.Busy, s.state.Message = true, ""
	s.workers.Add(1)
	go func() {
		defer s.workers.Done()
		err := s.ctx.Err()
		if err == nil {
			err = validateSettings(value)
		}
		if err == nil {
			err = s.repo.Save(value)
		}
		s.post(func() {
			if s.ctx.Err() != nil {
				return
			}
			s.state.Busy = false
			if err != nil {
				s.state.Message = err.Error()
				zap.L().Error("save settings", zap.Error(err))
			} else {
				s.state.Value = value
			}
		})
	}()
}

func validateSettings(value model.Settings) error {
	switch value.Language {
	case "", "system", "zh-CN", "en":
	default:
		return fmt.Errorf("invalid language: %s", value.Language)
	}
	switch value.Theme {
	case "system", "light", "dark":
	default:
		return fmt.Errorf("invalid theme: %s", value.Theme)
	}
	if value.Log.MaxAge < 0 {
		return fmt.Errorf("log max_age must not be negative")
	}
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(value.Log.Level)); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	return nil
}
