package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"go.uber.org/zap"

	"go-desktop-template/internal/db"
	"go-desktop-template/internal/pkg/i18n"
	"go-desktop-template/internal/pkg/logger"
	"go-desktop-template/internal/repository"
	"go-desktop-template/internal/store"
)

// App owns the window, stores and resources for one application lifetime.
type App struct {
	translator         *i18n.Translator
	trayMenu           *mygo.Menu
	window             *mygo.Window
	tray               *mygo.Tray
	todos              *store.TodoStore
	settings           *store.SettingsStore
	database           *sql.DB
	todoRepository     *repository.TodoRepository
	settingsRepository *repository.SettingsRepository
	quitting           bool
}

func Run() (runErr error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("resolve user configuration directory: %w", err)
	}
	base = filepath.Join(base, "go-desktop-template")

	a := &App{}
	a.settingsRepository = repository.NewSettingsRepository(filepath.Join(base, "settings.json"))
	preferences, err := a.settingsRepository.Load()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	a.settings, err = store.NewSettingsStore(a.settingsRepository, preferences, a.post)
	if err != nil {
		return fmt.Errorf("validate settings: %w", err)
	}
	defer func() { a.settings.Close(); a.settings.Wait() }()
	log, closeLogger, err := logger.New(filepath.Join(base, "logs"), preferences.Log)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer func() { runErr = errors.Join(runErr, closeLogger()) }()
	restoreLogger := zap.ReplaceGlobals(log)
	defer restoreLogger()
	a.translator, err = i18n.New(preferences.Language)
	if err != nil {
		return err
	}
	dataPath := filepath.Join(base, "data.db")
	a.database, err = db.Open(dataPath)
	if err != nil {
		return err
	}
	defer a.database.Close()
	a.todoRepository = repository.NewTodoRepository(a.database)
	items, err := a.todoRepository.List(context.Background())
	if err != nil {
		return fmt.Errorf("load todos: %w", err)
	}
	a.todos = store.NewTodoStore(a.todoRepository, items, a.post)
	defer func() {
		a.todos.Close()
		a.todos.Wait()
	}()
	mygo.App.WhenReady(a.ready)
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) {
		a.quitting = true
	})
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) {
		a.todos.Close()
		a.settings.Close()
	})
	return mygo.App.Run()
}
