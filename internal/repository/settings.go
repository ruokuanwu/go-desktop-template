package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go-desktop-template/internal/model"
)

// SettingsRepository persists application settings as JSON.
type SettingsRepository struct {
	path string
	mu   sync.Mutex
}

func NewSettingsRepository(path string) *SettingsRepository {
	return &SettingsRepository{path: path}
}

func (s *SettingsRepository) Load() (model.Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value := model.Settings{Language: "system", Theme: "system", Log: model.LogSettings{MaxAge: 7, Level: "info"}}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return value, nil
	}
	if err != nil {
		return value, fmt.Errorf("read settings: %w", err)
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("decode settings: %w", err)
	}
	return value, nil
}

func (s *SettingsRepository) Save(value model.Settings) (saveErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	return nil
}
