package store

import (
	"context"

	"go-desktop-template/internal/pkg/task"

	"go-desktop-template/internal/config"
)

type ConfigState struct {
	Value       config.Config
	Busy        bool
	Message     string
	TrayMessage string
}

type Config struct {
	jobs  *task.Task
	save  func(config.Config) error
	state ConfigState
}

func NewConfig(jobs *task.Task, initial config.Config, save func(config.Config) error) *Config {
	return &Config{jobs: jobs, save: save, state: ConfigState{Value: initial}}
}

func (s *Config) State() ConfigState { return s.state }

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
	if s.state.Busy || s.jobs.Err() != nil {
		return
	}
	s.state.Busy, s.state.Message = true, ""
	s.jobs.Run(func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return s.save(config)
	}, func(err error) {
		s.state.Busy = false
		if err != nil {
			s.state.Message = err.Error()
			return
		}
		s.state.Value = config
	})
}
