package store

import (
	"errors"
	"testing"
	"time"

	"go-desktop-template/internal/model"
)

type settingsServiceFunc func(model.Settings) error

func (f settingsServiceFunc) Save(value model.Settings) error { return f(value) }

func nextSettingsUpdate(t *testing.T, updates <-chan func()) func() {
	t.Helper()
	select {
	case publish := <-updates:
		return publish
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for UI update")
		return nil
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
			updates := make(chan func(), 1)
			initial := model.Settings{Theme: "system"}
			var saved model.Settings
			service := settingsServiceFunc(func(value model.Settings) error {
				saved = value
				return tt.err
			})
			s := NewSettingsStore(service, func(fn func()) { updates <- fn }, initial)
			defer func() { s.Close(); s.Wait() }()
			s.SetTheme("dark")
			publish := nextSettingsUpdate(t, updates)
			if saved.Theme != "dark" || !s.State().Busy || s.State().Value != initial {
				t.Fatalf("unexpected state before publication: %+v", s.State())
			}
			publish()
			if s.State().Busy {
				t.Fatal("store is still busy after publication")
			}
			if tt.err != nil {
				if s.State().Value != initial || s.State().Message != tt.err.Error() {
					t.Fatalf("unexpected failure state: %+v", s.State())
				}
				return
			}
			if s.State().Value.Theme != "dark" || s.State().Message != "" {
				t.Fatalf("unexpected saved state: %+v", s.State())
			}
			s.SetCloseToTray(true)
			nextSettingsUpdate(t, updates)()
			if !s.State().Value.CloseToTray || s.State().Value.Theme != "dark" {
				t.Fatalf("unexpected tray state: %+v", s.State())
			}
		})
	}
}

func TestSettingsCloseSuppressesQueuedPublication(t *testing.T) {
	updates := make(chan func(), 1)
	saves := 0
	initial := model.Settings{Theme: "system"}
	service := settingsServiceFunc(func(model.Settings) error {
		saves++
		return nil
	})
	s := NewSettingsStore(service, func(fn func()) { updates <- fn }, initial)
	defer func() { s.Close(); s.Wait() }()
	s.SetTheme("dark")
	publish := nextSettingsUpdate(t, updates)
	s.Close()
	s.Wait()
	publish()
	s.SetCloseToTray(true)
	if s.State().Value != initial || saves != 1 {
		t.Fatalf("closed store changed state or saved again: %+v, saves=%d", s.State(), saves)
	}
}
