package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/egoist/mygo"

	"go-desktop-template/internal/config"
	"go-desktop-template/internal/db"
	"go-desktop-template/internal/desktop"
	"go-desktop-template/internal/repository"
	"go-desktop-template/internal/service"
	"go-desktop-template/internal/view"
)

func Run() error {
	dir, err := config.DataDir()
	if err != nil {
		return err
	}
	database, err := db.Open(filepath.Join(dir, "todos.db"))
	if err != nil {
		return err
	}
	defer database.Close()
	todos := service.New(repository.New(database))
	items, err := todos.List(context.Background())
	if err != nil {
		return fmt.Errorf("load todos: %w", err)
	}
	preferences, err := view.LoadPreferences(dir)
	if err != nil {
		return fmt.Errorf("load preferences: %w", err)
	}
	screen := view.New(todos, items, preferences, dir)
	defer func() { screen.Cancel(); screen.Wait() }()
	mygo.App.WhenReady(func() { desktop.Open(screen) })
	return mygo.App.Run()
}
