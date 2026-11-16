package store

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-desktop-template/internal/model"
)

type settingsServiceFunc func(model.Settings) error

func (f settingsServiceFunc) Save(value model.Settings) error { return f(value) }

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
			initial := model.Settings{Theme: "system"}
			service := settingsServiceFunc(func(value model.Settings) error {
				saved <- value
				<-release
				return tt.err
			})
			s := NewSettingsStore(service, initial)
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
	initial := model.Settings{Theme: "system"}
	service := settingsServiceFunc(func(model.Settings) error {
		saves.Add(1)
		close(started)
		<-release
		return nil
	})
	s := NewSettingsStore(service, initial)
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
	s := NewSettingsStore(settingsServiceFunc(func(model.Settings) error { return nil }), model.Settings{Theme: "system"})
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
	s := NewSettingsStore(settingsServiceFunc(func(model.Settings) error { return nil }), model.Settings{})
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
