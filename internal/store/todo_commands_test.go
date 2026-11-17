package store

import (
	"context"
	"errors"
	"sync"
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

// uiQueue models asynchronous dispatch; tests execute callbacks on their own thread.
type uiQueue chan func()

func (q uiQueue) post(fn func()) { q <- fn }

func (q uiQueue) run(t *testing.T) {
	t.Helper()
	select {
	case fn := <-q:
		fn()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for UI callback")
	}
}

func TestTodoCommandsPublishAfterPersistence(t *testing.T) {
	q := make(uiQueue, 1)
	repo := &blockingTodoRepository{started: make(chan struct{}), release: make(chan struct{})}
	var unblock sync.Once
	s := NewTodoStore(repo, nil, q.post)
	defer func() { unblock.Do(func() { close(repo.release) }); s.Close(); s.Wait() }()
	state := s.State()
	added := false
	s.Add("created", func() {
		added = true
		if len(s.State().Items) != 1 {
			t.Error("callback ran before publication")
		}
	})
	if !state.Busy || len(state.Items) != 0 || added {
		t.Fatal("state changed before persistence")
	}
	s.Delete(2)
	if state.Message != "正在保存任务，请稍后重试。" {
		t.Fatal("busy command accepted")
	}
	unblock.Do(func() { close(repo.release) })
	s.workers.Wait()
	if !state.Busy || added || len(state.Items) != 0 {
		t.Fatal("worker changed UI state")
	}
	q.run(t)
	if state.Busy || len(state.Items) != 1 || state.Items[0].ID != 2 || !added || state.Message != "" {
		t.Fatalf("unexpected create state: %+v", state)
	}
	s.SetCompleted(2, true)
	if !state.Busy || state.Items[0].Completed {
		t.Fatal("completion changed before persistence")
	}
	q.run(t)
	if state.Busy || !state.Items[0].Completed {
		t.Fatal("completion was not published")
	}
	s.Delete(2)
	if !state.Busy || len(state.Items) != 1 {
		t.Fatal("item removed before persistence")
	}
	q.run(t)
	if len(state.Items) != 0 || state.Busy {
		t.Fatal("deletion was not published")
	}
	if s.State() != state {
		t.Fatal("state pointer changed")
	}
}

func TestTodoCommandFailurePreservesItems(t *testing.T) {
	for _, command := range []string{"add", "complete", "delete"} {
		t.Run(command, func(t *testing.T) {
			q := make(uiQueue, 1)
			wantErr := errors.New("write failed")
			s := NewTodoStore(&todoRepositoryStub{err: wantErr}, []model.Todo{{ID: 1}}, q.post)
			defer func() { s.Close(); s.Wait() }()
			called := false
			switch command {
			case "add":
				s.Add("created", func() { called = true })
			case "complete":
				s.SetCompleted(1, true)
			case "delete":
				s.Delete(1)
			}
			q.run(t)
			state := s.State()
			if called || state.Busy || state.Message != wantErr.Error() || len(state.Items) != 1 || state.Items[0].Completed {
				t.Fatalf("unexpected failure state: %+v", state)
			}
		})
	}
}

func TestTodoCloseSuppressesPublication(t *testing.T) {
	for _, queued := range []bool{false, true} {
		t.Run(map[bool]string{false: "in-flight", true: "queued"}[queued], func(t *testing.T) {
			q := make(uiQueue, 1)
			repo := &blockingTodoRepository{started: make(chan struct{}), release: make(chan struct{})}
			var unblock sync.Once
			s := NewTodoStore(repo, nil, q.post)
			defer func() { unblock.Do(func() { close(repo.release) }); s.Close(); s.Wait() }()
			called := false
			s.Add("created", func() { called = true })
			if queued {
				unblock.Do(func() { close(repo.release) })
				s.workers.Wait()
			}
			s.Close()
			unblock.Do(func() { close(repo.release) })
			s.Wait()
			q.run(t)
			before := *s.State()
			s.Add("ignored", nil)
			s.SetCompleted(2, true)
			s.Delete(2)
			if called || len(s.State().Items) != 0 || s.State().Message != before.Message {
				t.Fatal("closed store published a result or accepted a command")
			}
			if len(q) != 0 {
				t.Fatal("closed store queued work")
			}
		})
	}
}

func TestTodoStateOwnership(t *testing.T) {
	items := []model.Todo{{ID: 1, Title: "initial"}}
	q := make(uiQueue, 1)
	s := NewTodoStore(&todoRepositoryStub{}, items, q.post)
	defer func() { s.Close(); s.Wait() }()
	items[0].Title = "changed"
	if s.State().Items[0].Title != "initial" {
		t.Fatal("store shares initial items with caller")
	}
	state := s.State()
	state.Message = "UI change"
	if s.State() != state || s.State().Message != "UI change" {
		t.Fatal("State does not return shared UI state")
	}
}
