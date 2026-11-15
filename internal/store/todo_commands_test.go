package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-desktop-template/internal/model"
)

type todoServiceStub struct {
	err error
}

func (s *todoServiceStub) Create(context.Context, string) (model.Todo, error) {
	return model.Todo{ID: 2, Title: "created"}, s.err
}

func (s *todoServiceStub) SetCompleted(context.Context, int64, bool) error { return s.err }
func (s *todoServiceStub) Delete(context.Context, int64) error             { return s.err }

func nextTodoUpdate(t *testing.T, updates <-chan func()) func() {
	t.Helper()
	select {
	case publish := <-updates:
		return publish
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for UI update")
		return nil
	}
}

func TestTodoCommandsPublishAfterPersistence(t *testing.T) {
	updates := make(chan func(), 1)
	s := NewTodo(&todoServiceStub{}, func(fn func()) { updates <- fn }, nil)
	defer func() { s.Close(); s.Wait() }()
	added := false
	s.Add("created", func() { added = true })
	publish := nextTodoUpdate(t, updates)
	if !s.State().Busy || len(s.State().Items) != 0 || added {
		t.Fatal("state changed before UI publication")
	}
	publish()
	if s.State().Busy || len(s.State().Items) != 1 || s.State().Items[0].ID != 2 || !added {
		t.Fatalf("unexpected create state: %+v", s.State())
	}
	s.SetCompleted(2, true)
	publish = nextTodoUpdate(t, updates)
	if s.State().Items[0].Completed {
		t.Fatal("completion changed before publication")
	}
	publish()
	if !s.State().Items[0].Completed {
		t.Fatal("completion was not published")
	}
	s.Delete(2)
	publish = nextTodoUpdate(t, updates)
	if len(s.State().Items) != 1 {
		t.Fatal("item removed before publication")
	}
	publish()
	if len(s.State().Items) != 0 || s.State().Busy {
		t.Fatal("deletion was not published")
	}
}

func TestTodoCommandFailurePreservesItems(t *testing.T) {
	wantErr := errors.New("write failed")
	updates := make(chan func(), 1)
	s := NewTodo(&todoServiceStub{err: wantErr}, func(fn func()) { updates <- fn }, []model.Todo{{ID: 1}})
	defer func() { s.Close(); s.Wait() }()
	called := false
	s.Add("created", func() { called = true })
	nextTodoUpdate(t, updates)()
	if called || s.State().Busy || s.State().Message != wantErr.Error() || len(s.State().Items) != 1 {
		t.Fatalf("unexpected failure state: %+v", s.State())
	}
}

func TestTodoCloseSuppressesQueuedPublication(t *testing.T) {
	updates := make(chan func(), 1)
	s := NewTodo(&todoServiceStub{}, func(fn func()) { updates <- fn }, nil)
	defer func() { s.Close(); s.Wait() }()
	called := false
	s.Add("created", func() { called = true })
	publish := nextTodoUpdate(t, updates)
	s.Close()
	publish()
	if called || len(s.State().Items) != 0 {
		t.Fatal("closed store published a result")
	}
}
