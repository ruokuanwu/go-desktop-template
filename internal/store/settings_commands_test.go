package store

import (
	"errors"
	"os"
	"path/filepath"

	"go-desktop-template/internal/repository"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-desktop-template/internal/model"
)

type settingsRepositoryFunc func(model.Settings) error

func (f settingsRepositoryFunc) Save(value model.Settings) error { return f(value) }

func nextSettingsState(t *testing.T, updates <-chan SettingsState) SettingsState {
	t.Helper()
	select {
	case state := <-updates:
		return state
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for state notification")
		return SettingsState{}
	}
}

func TestSettingsCommandsPublishAfterPersistence(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "success"},
		{name: "failure", err: errors.New("write failed")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			updates := make(chan SettingsState, 4)
			saved := make(chan model.Settings, 1)
			release := make(chan struct{})
			var unblock sync.Once
			initial := model.Settings{Theme: "system", Log: model.LogSettings{Level: "info"}}
			repo := settingsRepositoryFunc(func(value model.Settings) error {
				saved <- value
				<-release
				return tt.err
			})
			s := newSettingsStore(t, repo, initial)
			defer func() {
				unblock.Do(func() { close(release) })
				s.Close()
				s.Wait()
			}()
			unsubscribe := s.Subscribe(func(state SettingsState) { updates <- state })
			defer unsubscribe()
			s.SetTheme("dark")
			state := nextSettingsState(t, updates)
			if !state.Busy || state.Value != initial {
				t.Fatalf("unexpected state before persistence: %+v", state)
			}
			s.SetCloseToTray(true)
			unblock.Do(func() { close(release) })
			state = nextSettingsState(t, updates)
			if value := <-saved; value.Theme != "dark" || value.CloseToTray {
				t.Fatalf("unexpected saved settings: %+v", value)
			}
			if state.Busy {
				t.Fatal("store is still busy after persistence")
			}
			if tt.err != nil {
				if state.Value != initial || state.Message != tt.err.Error() {
					t.Fatalf("unexpected failure state: %+v", state)
				}
				return
			}
			if state.Value.Theme != "dark" || state.Message != "" {
				t.Fatalf("unexpected saved state: %+v", state)
			}
			s.SetCloseToTray(true)
			nextSettingsState(t, updates)
			state = nextSettingsState(t, updates)
			if !state.Value.CloseToTray || state.Value.Theme != "dark" {
				t.Fatalf("unexpected tray state: %+v", state)
			}
		})
	}
}

func TestSettingsCloseSuppressesPublication(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var unblock sync.Once
	var saves, notifications atomic.Int32
	initial := model.Settings{Theme: "system", Log: model.LogSettings{Level: "info"}}
	repo := settingsRepositoryFunc(func(model.Settings) error {
		saves.Add(1)
		close(started)
		<-release
		return nil
	})
	s := newSettingsStore(t, repo, initial)
	defer func() {
		unblock.Do(func() { close(release) })
		s.Close()
		s.Wait()
	}()
	s.Subscribe(func(SettingsState) { notifications.Add(1) })
	s.SetTheme("dark")
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for save")
	}
	s.Close()
	s.Subscribe(func(SettingsState) { notifications.Add(1) })()
	unblock.Do(func() { close(release) })
	s.Wait()
	s.SetCloseToTray(true)
	s.SetTrayError(errors.New("tray failed"))
}

func TestSettingsSubscriptionsAndSnapshot(t *testing.T) {
	s := newSettingsStore(t, settingsRepositoryFunc(func(model.Settings) error { return nil }), model.Settings{Theme: "system", Log: model.LogSettings{Level: "info"}})
	defer func() { s.Close(); s.Wait() }()
	var first, second int
	unsubscribe := s.Subscribe(func(state SettingsState) {
		state.Value.Theme = "subscriber change"
		first++
		_ = s.State()
	})
	s.Subscribe(func(SettingsState) { second++ })
	if first != 0 || second != 0 {
		t.Fatal("subscription sent an initial notification")
	}
	s.SetTrayError(errors.New("tray failed"))
	unsubscribe()
	unsubscribe()
	s.SetTrayError(errors.New("another failure"))
	if first != 1 || second != 2 {
		t.Fatalf("unexpected notifications after unsubscribe: %d, %d", first, second)
	}
	snapshot := s.State()
	snapshot.Value.Theme = "dark"
	if s.State().Value.Theme != "system" {
		t.Fatal("modifying snapshot changed store state")
	}
}

func TestSettingsConcurrentCommandsAndClose(t *testing.T) {
	s := newSettingsStore(t, settingsRepositoryFunc(func(model.Settings) error { return nil }), model.Settings{Theme: "system", Log: model.LogSettings{Level: "info"}})
	var callers sync.WaitGroup
	for range 16 {
		callers.Go(func() {
			for range 50 {
				unsubscribe := s.Subscribe(func(SettingsState) { _ = s.State() })
				s.SetTheme("dark")
				s.SetCloseToTray(true)
				s.SetTrayError(errors.New("tray failed"))
				unsubscribe()
			}
		})
	}
	callers.Go(func() { s.Close(); s.Wait() })
	callers.Wait()
	s.Close()
	s.Wait()
}

func newSettingsStore(t *testing.T, repo SettingsRepository, initial model.Settings) *SettingsStore {
	t.Helper()
	s, err := NewSettingsStore(repo, initial)
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
			if s, err := NewSettingsStore(repo, value); err == nil {
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
			s := newSettingsStore(t, repo, initial)
			defer func() { s.Close(); s.Wait() }()
			s.saveValue(func(settings *model.Settings) { *settings = value })
			s.workers.Wait()
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
