package store

import (
	"context"

	"go-desktop-template/internal/pkg/task"

	"go-desktop-template/internal/config"
)

type PreferenceState struct {
	Value   config.Preferences
	Busy    bool
	Message string
}

type Preferences struct {
	jobs  *task.Task
	save  func(config.Preferences) error
	state PreferenceState
}

func NewPreferences(jobs *task.Task, initial config.Preferences, save func(config.Preferences) error) *Preferences {
	return &Preferences{jobs: jobs, save: save, state: PreferenceState{Value: initial}}
}

func (s *Preferences) State() PreferenceState { return s.state }

func (s *Preferences) SetAppearance(appearance string) {
	if s.state.Busy || s.jobs.Err() != nil {
		return
	}
	s.state.Busy, s.state.Message = true, ""
	p := config.Preferences{Appearance: appearance}
	s.jobs.Run(func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return s.save(p)
	}, func(err error) {
		s.state.Busy = false
		if err != nil {
			s.state.Message = err.Error()
			return
		}
		s.state.Value = p
	})
}
