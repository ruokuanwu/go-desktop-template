package config

import (
	"os"
	"path/filepath"
)

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
