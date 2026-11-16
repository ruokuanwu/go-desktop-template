package repository

import (
	"os"
	"path/filepath"
	"testing"

	"go-desktop-template/internal/model"
)

func TestSettingsLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	s := NewSettingsRepository(filepath.Join(dir, "settings.json"))
	value, err := s.Load()
	want := model.Settings{Theme: "system", Log: model.LogSettings{MaxAge: 7, Level: "info"}}
	if err != nil || value != want {
		t.Fatalf("load defaults: value=%+v, err=%v", value, err)
	}
}

func TestSettingsPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	s := NewSettingsRepository(path)
	want := model.Settings{Theme: "light", CloseToTray: true, Log: model.LogSettings{MaxAge: 14, Level: "debug"}}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	value, err := NewSettingsRepository(path).Load()
	if err != nil || value != want {
		t.Fatalf("reload settings: value=%+v, err=%v", value, err)
	}

}

func TestSettingsLoadExistingPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"theme":"dark","close_to_tray":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := NewSettingsRepository(path).Load()
	want := model.Settings{Theme: "dark", CloseToTray: true, Log: model.LogSettings{MaxAge: 7, Level: "info"}}
	if err != nil || value != want {
		t.Fatalf("load existing preferences: value=%+v, err=%v", value, err)
	}
}

func TestSettingsLoadRejectsInvalidFile(t *testing.T) {
	for _, data := range []string{`{`} {
		t.Run(data, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "settings.json")
			if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := NewSettingsRepository(path).Load(); err == nil {
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
	if err := NewSettingsRepository(path).Save(model.Settings{Theme: "system"}); err == nil {
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
