package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"go-desktop-template/internal/model"
)

type Repository interface {
	List(ctx context.Context) ([]model.Todo, error)
	Create(ctx context.Context, title string) (model.Todo, error)
	SetCompleted(ctx context.Context, id int64, completed bool) error
	Delete(ctx context.Context, id int64) error
}

type Todos struct {
	repo Repository
}

func New(repo Repository) *Todos {
	return &Todos{repo: repo}
}

func (s *Todos) List(ctx context.Context) ([]model.Todo, error) {
	return s.repo.List(ctx)
}

func (s *Todos) Create(ctx context.Context, title string) (model.Todo, error) {
	title = strings.TrimSpace(title)
	length := utf8.RuneCountInString(title)
	if !utf8.ValidString(title) || length < 1 || length > 200 {
		return model.Todo{}, errors.New("todo title must contain 1–200 Unicode characters")
	}
	return s.repo.Create(ctx, title)
}

func (s *Todos) SetCompleted(ctx context.Context, id int64, completed bool) error {
	return s.repo.SetCompleted(ctx, id, completed)
}

func (s *Todos) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
