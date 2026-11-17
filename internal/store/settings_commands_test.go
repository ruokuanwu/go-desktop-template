package store

import (
	"errors"
	"os"
	"path/filepath"

	"go-desktop-template/internal/repository"
	"sync"
	"testing"

	"go-desktop-template/internal/model"
)

type settingsRepositoryFunc func(model.Settings) error

func (f settingsRepositoryFunc) Save(value model.Settings) error { return f(value) }

func TestSettingsCommandsPublishAfterPersistence(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{{name: "success"}, {name: "failure", err: errors.New("write failed")}} {
		t.Run(tt.name, func(t *testing.T) {
			q := make(uiQueue, 1)
			saved := make(chan model.Settings, 1)
			release := make(chan struct{})
			var unblock sync.Once
			initial := model.Settings{Theme: "system", Log: model.LogSettings{Level: "info"}}
			repo := settingsRepositoryFunc(func(value model.Settings) error { saved <- value; <-release; return tt.err })
			s := newSettingsStore(t, repo, initial, q.post)
			defer func() { unblock.Do(func() { close(release) }); s.Close(); s.Wait() }()
			state := s.State()
			s.SetTheme("dark")
			if !state.Busy || state.Value != initial {
				t.Fatalf("unexpected state before persistence: %+v", state)
			}
			s.SetCloseToTray(true)
			unblock.Do(func() { close(release) })
			s.workers.Wait()
			if state.Value != initial || !state.Busy {
				t.Fatal("worker changed UI state")
			}
			q.run(t)
			if value := <-saved; value.Theme != "dark" || value.CloseToTray {
				t.Fatalf("unexpected saved settings: %+v", value)
			}
			if state.Busy {
				t.Fatal("still busy after persistence")
			}
			if tt.err != nil {
				if state.Value != initial || state.Message != tt.err.Error() {
					t.Fatalf("unexpected failure: %+v", state)
				}
				return
			}
			if state.Value.Theme != "dark" || state.Message != "" {
				t.Fatalf("unexpected saved state: %+v", state)
			}
			s.SetCloseToTray(true)
			q.run(t)
			if !state.Value.CloseToTray || state.Value.Theme != "dark" {
				t.Fatalf("unexpected tray state: %+v", state)
			}
			if state != s.State() {
				t.Fatal("state pointer changed")
			}
		})
	}
}

func TestSettingsCloseSuppressesPublication(t *testing.T) {
	for _, queued := range []bool{false, true} {
		t.Run(map[bool]string{false: "in-flight", true: "queued"}[queued], func(t *testing.T) {
			q := make(uiQueue, 1)
			release := make(chan struct{})
			var unblock sync.Once
			initial := model.Settings{Theme: "system", Log: model.LogSettings{Level: "info"}}
			s := newSettingsStore(t, settingsRepositoryFunc(func(model.Settings) error { <-release; return nil }), initial, q.post)
			defer func() { unblock.Do(func() { close(release) }); s.Close(); s.Wait() }()
			s.SetTheme("dark")
			if queued {
				unblock.Do(func() { close(release) })
				s.workers.Wait()
			}
			s.Close()
			unblock.Do(func() { close(release) })
			s.Wait()
			q.run(t)
			before := *s.State()
			s.SetCloseToTray(true)
			s.SetTheme("light")
			if *s.State() != before || s.State().Value != initial || len(q) != 0 {
				t.Fatal("closed store changed state or queued work")
			}
		})
	}
}

func newSettingsStore(t *testing.T, repo SettingsRepository, initial model.Settings, post func(func())) *SettingsStore {
	t.Helper()
	s, err := NewSettingsStore(repo, initial, post)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSettingsLoadRejectsInvalidFile(t *testing.T) {
	for _, data := range []string{`{"theme":"invalid"}`, `{"log":{"max_age":-1}}`, `{"log":{"level":"invalid"}}`} {
		t.Run(data, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.json")
			if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}
			repo := repository.NewSettingsRepository(path)
			value, err := repo.Load()
			if err != nil {
				t.Fatal(err)
			}
			if s, err := NewSettingsStore(repo, value, make(uiQueue, 1).post); err == nil {
				s.Close()
				s.Wait()
				t.Fatal("invalid settings accepted")
			}
		})
	}
}

func TestSettingsSaveRejectsInvalidSettings(t *testing.T) {
	for _, value := range []model.Settings{
		{Theme: "invalid", Log: model.LogSettings{Level: "info"}},
		{Theme: "system", Log: model.LogSettings{MaxAge: -1, Level: "info"}},
		{Theme: "system", Log: model.LogSettings{MaxAge: 7, Level: "invalid"}},
	} {
		t.Run(value.Theme+value.Log.Level, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.json")
			repo := repository.NewSettingsRepository(path)
			initial, err := repo.Load()
			if err != nil {
				t.Fatal(err)
			}
			if err := repo.Save(initial); err != nil {
				t.Fatal(err)
			}
			q := make(uiQueue, 1)
			s := newSettingsStore(t, repo, initial, q.post)
			defer func() { s.Close(); s.Wait() }()
			s.saveValue(func(settings *model.Settings) { *settings = value })
			q.run(t)
			state := s.State()
			if state.Busy || state.Message == "" || state.Value != initial {
				t.Fatalf("invalid save changed state: %+v", state)
			}
			persisted, err := repo.Load()
			if err != nil || persisted != initial {
				t.Fatalf("invalid save changed persisted settings: value=%+v, err=%v", persisted, err)
			}
		})
	}
}
