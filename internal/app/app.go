package app

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"go-desktop-template/internal/config"
	"go-desktop-template/internal/db"
	"go-desktop-template/internal/pkg/gtkfix"
	"go-desktop-template/internal/pkg/task"
	"go-desktop-template/internal/repository"
	"go-desktop-template/internal/service"
	"go-desktop-template/internal/store"
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
	todos := service.NewTodo(repository.NewTodo(database))
	items, err := todos.List(context.Background())
	if err != nil {
		return fmt.Errorf("load todos: %w", err)
	}
	preferences, err := config.LoadPreferences(dir)
	if err != nil {
		return fmt.Errorf("load preferences: %w", err)
	}
	var jobs *task.Task
	defer func() {
		if jobs != nil {
			jobs.Close()
			jobs.Wait()
		}
	}()
	mygo.App.WhenReady(func() {
		if err := gtkfix.TextColor(); err != nil {
			log.Printf("apply text color workaround: %v", err)
		}
		var window *mygo.Window
		jobs = task.New(func(publish func()) { window.Update(publish) })
		todoStore := store.NewTodo(todos, jobs, items)
		preferenceStore := store.NewPreferences(jobs, preferences, func(p config.Preferences) error {
			return config.SavePreferences(dir, p)
		})
		screen := view.New(todoStore, preferenceStore, dir)
		window = mygo.NewWindow(mygo.WindowOptions{
			Title: "Todo Desktop", Width: 1000, Height: 720,
			MinWidth: 720, MinHeight: 520,
			StateKey: "main", Content: ui.View(screen.Render),
		})
		window.OnClosed(jobs.Close)
	})
	return mygo.App.Run()
}
