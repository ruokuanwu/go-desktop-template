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
	value, err := s.Load()
	want := model.Settings{Theme: "system", Log: model.LogSettings{MaxAge: 7, Level: "info"}}
	if err != nil || value != want {
		t.Fatalf("load defaults: value=%+v, err=%v", value, err)
	}
}

func TestSettingsPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	s := NewSettings(path)
	want := model.Settings{Theme: "light", CloseToTray: true, Log: model.LogSettings{MaxAge: 14, Level: "debug"}}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	value, err := NewSettings(path).Load()
	if err != nil || value != want {
		t.Fatalf("reload settings: value=%+v, err=%v", value, err)
	}
	if err := s.Save(model.Settings{Theme: "invalid"}); err == nil {
		t.Fatal("invalid settings accepted")
	}
	value, err = s.Load()
	if err != nil || value != want {
		t.Fatalf("invalid save changed persisted settings: value=%+v, err=%v", value, err)
	}
}

func TestSettingsLoadExistingPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"theme":"dark","close_to_tray":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := NewSettings(path).Load()
	want := model.Settings{Theme: "dark", CloseToTray: true, Log: model.LogSettings{MaxAge: 7, Level: "info"}}
	if err != nil || value != want {
		t.Fatalf("load existing preferences: value=%+v, err=%v", value, err)
	}
}

func TestSettingsLoadRejectsInvalidFile(t *testing.T) {
	for _, data := range []string{`{`, `{"theme":"invalid"}`, `{"log":{"max_age":-1}}`, `{"log":{"level":"invalid"}}`} {
		t.Run(data, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "settings.json")
			if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := NewSettings(path).Load(); err == nil {
				t.Fatal("invalid file silently ignored")
			}
		})
	}
}

func TestSettingsSaveRejectsInvalidLogging(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	for _, log := range []model.LogSettings{{MaxAge: -1, Level: "info"}, {MaxAge: 7, Level: "invalid"}} {
		if err := NewSettings(path).Save(model.Settings{Theme: "system", Log: log}); err == nil {
			t.Fatalf("invalid logging settings accepted: %+v", log)
		}
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("invalid save created settings: %v", err)
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
