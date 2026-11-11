package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Theme string `json:"theme"`
}

func Load(dir string) (Config, error) {
	config := Config{Theme: "system"}
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if os.IsNotExist(err) {
		data, err = os.ReadFile(filepath.Join(dir, "preferences.json"))
	}
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}
	if err := validate(config); err != nil {
		return config, err
	}
	return config, nil
}

func Save(dir string, config Config) error {
	if err := validate(config); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	temp := filepath.Join(dir, "config.json.tmp")
	if err := os.WriteFile(temp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, filepath.Join(dir, "config.json"))
}

func validate(config Config) error {
	switch config.Theme {
	case "system", "light", "dark":
		return nil
	default:
		return fmt.Errorf("invalid theme: %s", config.Theme)
	}
}

func DataDir() (string, error) {
	if dir := os.Getenv("TODO_DATA_DIR"); dir != "" {
		return dir, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "go-desktop-template"), nil
}
