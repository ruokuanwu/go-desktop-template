package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go-desktop-template/internal/model"
)

var ErrNotFound = errors.New("todo not found")

type Todos struct {
	db *sql.DB
}

func New(db *sql.DB) *Todos {
	return &Todos{db: db}
}

func (r *Todos) List(ctx context.Context) ([]model.Todo, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, completed FROM todos ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := make([]model.Todo, 0)
	for rows.Next() {
		var todo model.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *Todos) Create(ctx context.Context, title string) (model.Todo, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO todos (title, completed) VALUES (?, ?)", title, false)
	if err != nil {
		return model.Todo{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Todo{}, err
	}
	return model.Todo{ID: id, Title: title, Completed: false}, nil
}

func (r *Todos) SetCompleted(ctx context.Context, id int64, completed bool) error {
	result, err := r.db.ExecContext(ctx, "UPDATE todos SET completed = ? WHERE id = ?", completed, id)
	return checkAffected(result, err, id)
}

func (r *Todos) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM todos WHERE id = ?", id)
	return checkAffected(result, err, id)
}

func checkAffected(result sql.Result, err error, id int64) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%w: %d", ErrNotFound, id)
	}
	return nil
}
