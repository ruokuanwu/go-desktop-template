package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap/zapcore"
)

type Config struct {
	DataDir string    `json:"data_dir"`
	Theme   string    `json:"theme"`
	Log     LogConfig `json:"log"`
}

type LogConfig struct {
	LogDir string `json:"log_dir"`
	MaxAge int    `json:"max_age"`
	Level  string `json:"level"`
}

func Load(path string) (Config, error) {
	config := Config{
		Theme: "system",
		Log:   LogConfig{LogDir: "logs", MaxAge: 7, Level: "info"},
	}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return config, err
	}
	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return config, err
		}
	}

	if !filepath.IsAbs(config.DataDir) {
		base, err := os.UserConfigDir()
		if err != nil {
			return config, err
		}
		config.DataDir = filepath.Join(base, "go-desktop-template", "data.db")
	}

	if !filepath.IsAbs(config.Log.LogDir) {
		base, err := os.UserConfigDir()
		if err != nil {
			return config, err
		}
		config.Log.LogDir = filepath.Join(base, "go-desktop-template", "logs")
	}

	if err := validate(config); err != nil {
		return config, err
	}
	return config, nil
}

func Save(path string, config Config) error {
	if err := validate(config); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func validate(config Config) error {
	if config.DataDir == "" {
		return fmt.Errorf("data directory must not be empty")
	}
	switch config.Theme {
	case "system", "light", "dark":
	default:
		return fmt.Errorf("invalid theme: %s", config.Theme)
	}
	if config.Log.LogDir == "" {
		return fmt.Errorf("log directory must not be empty")
	}
	if config.Log.MaxAge < 0 {
		return fmt.Errorf("log max_age must not be negative")
	}
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.Log.Level)); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	return nil
}
