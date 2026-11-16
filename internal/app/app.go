package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"go-desktop-template/internal/db"
	"go-desktop-template/internal/desktop"
	"go-desktop-template/internal/pkg/logger"
	"go-desktop-template/internal/repository"
	"go-desktop-template/internal/service"
	"go-desktop-template/internal/store"
)

func Run() (runErr error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("resolve user configuration directory: %w", err)
	}
	base = filepath.Join(base, "go-desktop-template")

	settings := service.NewSettingsService(filepath.Join(base, "settings.json"))
	preferences, err := settings.Load()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	log, closeLogger, err := logger.New(filepath.Join(base, "logs"), preferences.Log)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer func() { runErr = errors.Join(runErr, closeLogger()) }()
	restoreLogger := zap.ReplaceGlobals(log)
	defer restoreLogger()
	dataPath := filepath.Join(base, "data.db")
	database, err := db.Open(dataPath)
	if err != nil {
		return err
	}
	defer database.Close()
	todos := service.NewTodoService(repository.NewTodoRepository(database))
	items, err := todos.List(context.Background())
	if err != nil {
		return fmt.Errorf("load todos: %w", err)
	}
	todoStore := store.NewTodoStore(todos, items)
	settingsStore := store.NewSettingsStore(settings, preferences)
	defer func() {
		todoStore.Close()
		settingsStore.Close()
		todoStore.Wait()
		settingsStore.Wait()
	}()
	return desktop.Run(desktop.Options{
		Todos:    todoStore,
		Settings: settingsStore,
	})
}
