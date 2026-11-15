package store

import (
	"errors"
	"testing"
	"time"

	"go-desktop-template/internal/config"
)

func nextConfigUpdate(t *testing.T, updates <-chan func()) func() {
	t.Helper()
	select {
	case publish := <-updates:
		return publish
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for UI update")
		return nil
	}
}

func TestConfigCommandsPublishAfterPersistence(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "success"},
		{name: "failure", err: errors.New("write failed")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			updates := make(chan func(), 1)
			initial := config.Config{Theme: "system"}
			var saved config.Config
			s := NewConfig(func(fn func()) { updates <- fn }, initial, func(value config.Config) error {
				saved = value
				return tt.err
			})
			defer func() { s.Close(); s.Wait() }()
			s.SetTheme("dark")
			publish := nextConfigUpdate(t, updates)
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
			nextConfigUpdate(t, updates)()
			if !s.State().Value.CloseToTray || s.State().Value.Theme != "dark" {
				t.Fatalf("unexpected tray state: %+v", s.State())
			}
		})
	}
}

func TestConfigCloseSuppressesQueuedPublication(t *testing.T) {
	updates := make(chan func(), 1)
	saves := 0
	initial := config.Config{Theme: "system"}
	s := NewConfig(func(fn func()) { updates <- fn }, initial, func(config.Config) error {
		saves++
		return nil
	})
	defer func() { s.Close(); s.Wait() }()
	s.SetTheme("dark")
	publish := nextConfigUpdate(t, updates)
	s.Close()
	s.Wait()
	publish()
	s.SetCloseToTray(true)
	if s.State().Value != initial || saves != 1 {
		t.Fatalf("closed store changed state or saved again: %+v, saves=%d", s.State(), saves)
	}
}
