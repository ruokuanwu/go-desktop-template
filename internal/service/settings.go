package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go-desktop-template/internal/model"
)

// Settings persists user preferences separately from the startup configuration.
type Settings struct {
	path string
}

func NewSettings(path string) *Settings {
	return &Settings{path: path}
}

// Load imports preferences from the old combined config only if settings do not exist.
func (s *Settings) Load(legacyConfigPath string) (model.Settings, error) {
	value := model.Settings{Theme: "system"}
	data, err := os.ReadFile(s.path)
	legacy := os.IsNotExist(err)
	if legacy {
		data, err = os.ReadFile(legacyConfigPath)
	}
	if os.IsNotExist(err) {
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
	if legacy {
		if err := s.Save(value); err != nil {
			return value, fmt.Errorf("migrate settings: %w", err)
		}
	}
	return value, nil
}

func (s *Settings) Save(value model.Settings) (saveErr error) {
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
	return nil
}

func validateSettings(value model.Settings) error {
	switch value.Theme {
	case "system", "light", "dark":
		return nil
	default:
		return fmt.Errorf("invalid theme: %s", value.Theme)
	}
}
