package store

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-desktop-template/internal/model"
)

type todoRepositoryStub struct {
	err error
}

func (s *todoRepositoryStub) Create(context.Context, string) (model.Todo, error) {
	return model.Todo{ID: 2, Title: "created"}, s.err
}

func (s *todoRepositoryStub) SetCompleted(context.Context, int64, bool) error { return s.err }
func (s *todoRepositoryStub) Delete(context.Context, int64) error             { return s.err }

type blockingTodoRepository struct {
	todoRepositoryStub
	started chan struct{}
	release chan struct{}
}

func (s *blockingTodoRepository) Create(ctx context.Context, title string) (model.Todo, error) {
	close(s.started)
	<-s.release
	return s.todoRepositoryStub.Create(ctx, title)
}

func nextTodoState(t *testing.T, updates <-chan TodoState) TodoState {
	t.Helper()
	select {
	case state := <-updates:
		return state
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for state notification")
		return TodoState{}
	}
}

func TestTodoCommandsPublishAfterPersistence(t *testing.T) {
	updates := make(chan TodoState, 8)
	repo := &blockingTodoRepository{started: make(chan struct{}), release: make(chan struct{})}
	var unblock sync.Once
	s := NewTodoStore(repo, nil)
	defer func() { unblock.Do(func() { close(repo.release) }); s.Close(); s.Wait() }()
	s.Subscribe(func(state TodoState) { updates <- state })
	var added atomic.Bool
	s.Add("created", func() { added.Store(true); _ = s.State() })
	state := nextTodoState(t, updates)
	if !state.Busy || len(state.Items) != 0 || added.Load() {
		t.Fatal("state changed before persistence")
	}
	s.Delete(2)
	if state = nextTodoState(t, updates); !state.Busy || state.Message != "正在保存任务，请稍后重试。" {
		t.Fatalf("unexpected busy state: %+v", state)
	}
	unblock.Do(func() { close(repo.release) })
	state = nextTodoState(t, updates)
	if state.Busy || len(state.Items) != 1 || state.Items[0].ID != 2 || !added.Load() || state.Message != "" {
		t.Fatalf("unexpected create state: %+v", state)
	}
	s.SetCompleted(2, true)
	state = nextTodoState(t, updates)
	if !state.Busy || state.Items[0].Completed {
		t.Fatal("completion changed before persistence")
	}
	state = nextTodoState(t, updates)
	if state.Busy || !state.Items[0].Completed {
		t.Fatal("completion was not published")
	}
	s.Delete(2)
	state = nextTodoState(t, updates)
	if !state.Busy || len(state.Items) != 1 {
		t.Fatal("item removed before persistence")
	}
	state = nextTodoState(t, updates)
	if len(state.Items) != 0 || state.Busy {
		t.Fatal("deletion was not published")
	}
}

func TestTodoCommandFailurePreservesItems(t *testing.T) {
	for _, command := range []string{"add", "complete", "delete"} {
		t.Run(command, func(t *testing.T) {
			wantErr := errors.New("write failed")
			updates := make(chan TodoState, 2)
			s := NewTodoStore(&todoRepositoryStub{err: wantErr}, []model.Todo{{ID: 1}})
			defer func() { s.Close(); s.Wait() }()
			s.Subscribe(func(state TodoState) { updates <- state })
			var called atomic.Bool
			switch command {
			case "add":
				s.Add("created", func() { called.Store(true) })
			case "complete":
				s.SetCompleted(1, true)
			case "delete":
				s.Delete(1)
			}
			nextTodoState(t, updates)
			state := nextTodoState(t, updates)
			if called.Load() || state.Busy || state.Message != wantErr.Error() || len(state.Items) != 1 || state.Items[0].Completed {
				t.Fatalf("unexpected failure state: %+v", state)
			}
		})
	}
}

func TestTodoCloseSuppressesPublication(t *testing.T) {
	repo := &blockingTodoRepository{started: make(chan struct{}), release: make(chan struct{})}
	var unblock sync.Once
	s := NewTodoStore(repo, nil)
	defer func() { unblock.Do(func() { close(repo.release) }); s.Close(); s.Wait() }()
	var called atomic.Bool
	var notifications atomic.Int32
	s.Subscribe(func(TodoState) { notifications.Add(1) })
	s.Add("created", func() { called.Store(true) })
	select {
	case <-repo.started:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for save")
	}
	s.Close()
	s.Subscribe(func(TodoState) { notifications.Add(1) })()
	unblock.Do(func() { close(repo.release) })
	s.Wait()
	state := s.State()
	s.Add("ignored", nil)
	s.SetCompleted(2, true)
	s.Delete(2)
	if called.Load() || len(s.State().Items) != 0 || notifications.Load() != 1 || s.State().Message != state.Message {
		t.Fatal("closed store published a result or accepted a command")
	}
}

func TestTodoSnapshotsAndSubscriptions(t *testing.T) {
	items := []model.Todo{{ID: 1, Title: "initial"}}
	s := NewTodoStore(&todoRepositoryStub{}, items)
	defer func() { s.Close(); s.Wait() }()
	items[0].Title = "changed"
	snapshot := s.State()
	snapshot.Items[0].Title = "changed again"
	if s.State().Items[0].Title != "initial" {
		t.Fatal("store shares mutable items with caller")
	}
	var first, second atomic.Int32
	unsubscribe := s.Subscribe(func(state TodoState) {
		first.Add(1)
		state.Items[0].Title = "subscriber change"
		_ = s.State()
	})
	s.Subscribe(func(state TodoState) {
		second.Add(1)
		if len(state.Items) != 0 && state.Items[0].Title != "initial" {
			t.Error("subscribers share mutable items")
		}
	})
	if first.Load() != 0 || second.Load() != 0 {
		t.Fatal("unexpected initial notification")
	}
	s.SetCompleted(1, true)
	// Wait for this operation before checking notification counts.
	s.workers.Wait()
	if first.Load() != 2 || second.Load() != 2 {
		t.Fatal("missing notifications")
	}
	if s.State().Items[0].Title != "initial" {
		t.Fatal("subscriber changed store state")
	}
	unsubscribe()
	unsubscribe()
	s.Delete(1)
	s.workers.Wait()
	if first.Load() != 2 || second.Load() != 4 {
		t.Fatal("unsubscribe did not remove only its subscriber")
	}
	if snapshot.Items[0].Completed || snapshot.Items[0].Title != "changed again" {
		t.Fatal("later mutations changed a previous snapshot")
	}
}

func TestTodoConcurrentCommandsAndClose(t *testing.T) {
	s := NewTodoStore(&todoRepositoryStub{}, []model.Todo{{ID: 1}})
	var callers sync.WaitGroup
	for range 16 {
		callers.Go(func() {
			for range 50 {
				unsubscribe := s.Subscribe(func(TodoState) { _ = s.State() })
				s.Add("created", nil)
				s.SetCompleted(1, true)
				s.Delete(2)
				_ = s.State()
				unsubscribe()
			}
		})
	}
	callers.Go(func() { s.Close(); s.Wait() })
	callers.Wait()
	s.Close()
	s.Wait()
}
