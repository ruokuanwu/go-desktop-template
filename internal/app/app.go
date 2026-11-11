package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"go.uber.org/zap"

	"go-desktop-template/internal/config"
	"go-desktop-template/internal/db"
	"go-desktop-template/internal/pkg/gtkfix"
	"go-desktop-template/internal/pkg/logger"
	"go-desktop-template/internal/pkg/task"
	"go-desktop-template/internal/repository"
	"go-desktop-template/internal/service"
	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view"
)

func Run(configPath string) (runErr error) {
	initialConfig, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	dir := filepath.Dir(initialConfig.DataDir)
	log, closeLogger, err := logger.New(initialConfig.Log)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer func() { runErr = errors.Join(runErr, closeLogger()) }()
	log.Error("application starting", zap.String("data_dir", dir))
	database, err := db.Open(initialConfig.DataDir)
	if err != nil {
		return err
	}
	defer database.Close()
	todos := service.NewTodo(repository.NewTodo(database))
	items, err := todos.List(context.Background())
	if err != nil {
		return fmt.Errorf("load todos: %w", err)
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
			log.Warn("apply text color workaround", zap.Error(err))
		}
		var window *mygo.Window
		jobs = task.New(func(publish func()) { window.Update(publish) })
		todoStore := store.NewTodo(todos, jobs, items)
		configStore := store.NewConfig(jobs, initialConfig, func(value config.Config) error {
			return config.Save(configPath, value)
		})
		screen := view.New(todoStore, configStore, dir)
		window = mygo.NewWindow(mygo.WindowOptions{
			Title: "Todo Desktop", Width: 1000, Height: 720,
			MinWidth: 720, MinHeight: 520,
			StateKey: "main", Content: ui.View(screen.Render),
		})
		window.OnClosed(jobs.Close)
	})
	return mygo.App.Run()
}
