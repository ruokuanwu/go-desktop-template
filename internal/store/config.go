package store

import (
	"context"
	"sync"

	"go-desktop-template/internal/config"
)

type ConfigState struct {
	Value       config.Config
	Busy        bool
	Message     string
	TrayMessage string
}

type Config struct {
	save    func(config.Config) error
	update  func(func())
	ctx     context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
	state   ConfigState
}

func NewConfig(update func(func()), initial config.Config, save func(config.Config) error) *Config {
	ctx, cancel := context.WithCancel(context.Background())
	return &Config{save: save, update: update, ctx: ctx, cancel: cancel, state: ConfigState{Value: initial}}
}

func (s *Config) State() ConfigState { return s.state }

// Close stops accepting commands and cancels workers. Call it on the UI thread.
func (s *Config) Close() { s.cancel() }

// Wait waits for workers, not queued UI callbacks. Call it after Close, off the UI thread.
func (s *Config) Wait() { s.workers.Wait() }

func (s *Config) SetTrayError(err error) {
	s.state.TrayMessage = "系统托盘不可用，关闭窗口将退出程序：" + err.Error()
}

func (s *Config) SetTheme(theme string) {
	value := s.state.Value
	value.Theme = theme
	s.saveValue(value)
}

func (s *Config) SetCloseToTray(enabled bool) {
	value := s.state.Value
	value.CloseToTray = enabled
	s.saveValue(value)
}

func (s *Config) saveValue(config config.Config) {
	if s.state.Busy || s.ctx.Err() != nil {
		return
	}
	s.state.Busy, s.state.Message = true, ""
	s.workers.Add(1)
	go func() {
		defer s.workers.Done()
		err := s.ctx.Err()
		if err == nil {
			err = s.save(config)
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
			s.state.Value = config
		})
	}()
}
