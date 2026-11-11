package store

import (
	"context"

	"go-desktop-template/internal/pkg/task"

	"go-desktop-template/internal/config"
)

type ConfigState struct {
	Value   config.Config
	Busy    bool
	Message string
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

func (s *Config) SetTheme(theme string) {
	if s.state.Busy || s.jobs.Err() != nil {
		return
	}
	s.state.Busy, s.state.Message = true, ""
	config := s.state.Value
	config.Theme = theme
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
