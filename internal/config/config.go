package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Preferences struct {
	Appearance string `json:"appearance"`
}

func LoadPreferences(dir string) (Preferences, error) {
	p := Preferences{Appearance: "system"}
	data, err := os.ReadFile(filepath.Join(dir, "preferences.json"))
	if os.IsNotExist(err) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	if err := validatePreferences(p); err != nil {
		return p, err
	}
	return p, nil
}

func SavePreferences(dir string, p Preferences) error {
	if err := validatePreferences(p); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	temp := filepath.Join(dir, "preferences.json.tmp")
	if err := os.WriteFile(temp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, filepath.Join(dir, "preferences.json"))
}

func validatePreferences(p Preferences) error {
	switch p.Appearance {
	case "system", "light", "dark":
		return nil
	default:
		return fmt.Errorf("invalid appearance: %s", p.Appearance)
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
