package app

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go-desktop-template/internal/config"
	"go-desktop-template/internal/db"
	"go-desktop-template/internal/desktop"
	"go-desktop-template/internal/pkg/logger"
	"go-desktop-template/internal/pkg/task"
	"go-desktop-template/internal/repository"
	"go-desktop-template/internal/service"
	"go-desktop-template/internal/store"
)

func Run(configPath string) (runErr error) {
	initialConfig, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	log, closeLogger, err := logger.New(initialConfig.Log)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer func() { runErr = errors.Join(runErr, closeLogger()) }()
	restoreLogger := zap.ReplaceGlobals(log)
	defer restoreLogger()
	zap.L().Error("application starting", zap.String("data_dir", initialConfig.DataDir))
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
	var update func(func())
	jobs := task.New(func(publish func()) { update(publish) })
	defer func() {
		jobs.Close()
		jobs.Wait()
	}()
	todoStore := store.NewTodo(todos, jobs, items)
	configStore := store.NewConfig(jobs, initialConfig, func(value config.Config) error {
		return config.Save(configPath, value)
	})
	return desktop.Run(desktop.Options{
		Todos:  todoStore,
		Config: configStore,
		Lifecycle: desktop.Lifecycle{
			Ready:  func(publish func(func())) { update = publish },
			Closed: jobs.Close,
		},
	})
}
