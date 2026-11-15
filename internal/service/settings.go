package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap/zapcore"

	"go-desktop-template/internal/model"
)

// Settings persists application settings and provides the last committed snapshot.
type Settings struct {
	path       string
	operations sync.Mutex
	mu         sync.RWMutex
	current    model.Settings
	loaded     bool
}

func NewSettings(path string) *Settings {
	return &Settings{path: path}
}

func (s *Settings) Load() (model.Settings, error) {
	s.operations.Lock()
	defer s.operations.Unlock()

	value := model.Settings{Theme: "system", Log: model.LogSettings{MaxAge: 7, Level: "info"}}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		s.publish(value)
		return value, nil
	}
	if err != nil {
		return value, fmt.Errorf("read settings: %w", err)
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("decode settings: %w", err)
	}
	if err := validateSettings(value); err != nil {
		return value, err
	}
	s.publish(value)
	return value, nil
}

// Current returns a copy of the last successfully loaded or saved settings.
// It panics if neither Load nor Save has succeeded.
func (s *Settings) Current() model.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.loaded {
		panic("settings: no configuration loaded")
	}
	return s.current
}

func (s *Settings) publish(value model.Settings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current, s.loaded = value, true
}

func (s *Settings) Save(value model.Settings) (saveErr error) {
	s.operations.Lock()
	defer s.operations.Unlock()

	if err := validateSettings(value); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create settings directory: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(s.path), ".settings-*.tmp")
	if err != nil {
		return fmt.Errorf("create settings temporary file: %w", err)
	}
	defer func() {
		if saveErr != nil {
			if err := os.Remove(file.Name()); err != nil && !os.IsNotExist(err) {
				saveErr = fmt.Errorf("%w; remove temporary settings: %v", saveErr, err)
			}
		}
	}()
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write settings: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close settings: %w", closeErr)
	}
	if err := os.Rename(file.Name(), s.path); err != nil {
		return fmt.Errorf("replace settings: %w", err)
	}
	s.publish(value)
	return nil
}

func validateSettings(value model.Settings) error {
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
