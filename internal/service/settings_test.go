package service

import (
	"os"
	"path/filepath"
	"testing"

	"go-desktop-template/internal/model"
)

func TestSettingsLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	s := NewSettings(filepath.Join(dir, "settings.json"))
	value, err := s.Load(filepath.Join(dir, "config.json"))
	if err != nil || value != (model.Settings{Theme: "system"}) {
		t.Fatalf("load defaults: value=%+v, err=%v", value, err)
	}
}

func TestSettingsMigrationAndPersistence(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	legacy := `{"data_dir":"original.db","log":{"level":"debug"},"theme":"dark","close_to_tray":true}`
	if err := os.WriteFile(configPath, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "settings.json")
	s := NewSettings(path)
	want := model.Settings{Theme: "dark", CloseToTray: true}
	value, err := s.Load(configPath)
	if err != nil || value != want {
		t.Fatalf("migrate: value=%+v, err=%v", value, err)
	}
	persisted, err := os.ReadFile(path)
	if err != nil || string(persisted) != "{\n  \"theme\": \"dark\",\n  \"close_to_tray\": true\n}" {
		t.Fatalf("unexpected migrated settings: %s, err=%v", persisted, err)
	}
	want.Theme = "light"
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	value, err = NewSettings(path).Load(configPath)
	if err != nil || value != want {
		t.Fatalf("reload should prefer settings over legacy config: value=%+v, err=%v", value, err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil || string(data) != legacy {
		t.Fatalf("startup config changed: %s, err=%v", data, err)
	}
	if err := s.Save(model.Settings{Theme: "invalid"}); err == nil {
		t.Fatal("invalid settings accepted")
	}
	value, err = s.Load(configPath)
	if err != nil || value != want {
		t.Fatalf("invalid save changed persisted settings: value=%+v, err=%v", value, err)
	}
}

func TestSettingsLoadRejectsInvalidFile(t *testing.T) {
	for _, data := range []string{`{`, `{"theme":"invalid"}`} {
		t.Run(data, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "settings.json")
			if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := NewSettings(path).Load(filepath.Join(dir, "config.json")); err == nil {
				t.Fatal("invalid file silently ignored")
			}
		})
	}
}

func TestSettingsSaveFailureCleansTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := NewSettings(path).Save(model.Settings{Theme: "system"}); err == nil {
		t.Fatal("save to directory succeeded")
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name() != "settings.json" {
		t.Fatalf("temporary files not cleaned up: %v", files)
	}
}
