package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIgnoresLegacySettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{"theme":"invalid","close_to_tray":true}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := Load(path)
	if err != nil {
		t.Fatalf("startup config must not validate user preferences: %v", err)
	}
	if value.DataDir == "" || value.Log.Level != "info" {
		t.Fatalf("startup defaults lost: %+v", value)
	}
	persisted, err := os.ReadFile(path)
	if err != nil || string(persisted) != string(data) {
		t.Fatalf("loading config changed the file: %s, err=%v", persisted, err)
	}
}
