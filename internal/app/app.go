package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"go-desktop-template/internal/db"
	"go-desktop-template/internal/pkg/logger"
	"go-desktop-template/internal/repository"
	"go-desktop-template/internal/store"
	"go-desktop-template/internal/ui"
)

// App owns the desktop, stores and persistence resources for one application lifetime.
type App struct {
	desktop  *ui.Desktop
	todos    *store.TodoStore
	settings *store.SettingsStore
	database *sql.DB
	closeLog func() error
}

func Run() (runErr error) {
	a, err := newApp()
	if err != nil {
		return err
	}
	defer func() { runErr = errors.Join(runErr, a.close()) }()
	return a.desktop.Run(a.stop)
}

func newApp() (*App, error) {
	var desktop *ui.Desktop
	// Stores only post results from commands issued after the desktop opens.
	a, err := openResources(func(fn func()) { desktop.Post(fn) })
	if err != nil {
		return nil, err
	}
	desktop = ui.NewDesktop(a.todos, a.settings)
	a.desktop = desktop
	return a, nil
}

func openResources(post func(func())) (_ *App, openErr error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user configuration directory: %w", err)
	}
	base = filepath.Join(base, "go-desktop-template")

	a := &App{}
	defer func() {
		if openErr != nil {
			openErr = errors.Join(openErr, a.close())
		}
	}()

	settingsRepo := repository.NewSettingsRepository(filepath.Join(base, "settings.json"))
	preferences, err := settingsRepo.Load()
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}
	a.settings, err = store.NewSettingsStore(settingsRepo, preferences, post)
	if err != nil {
		return nil, fmt.Errorf("validate settings: %w", err)
	}
	log, closeLogger, err := logger.New(filepath.Join(base, "logs"), preferences.Log)
	if err != nil {
		return nil, fmt.Errorf("initialize logger: %w", err)
	}
	restoreLogger := zap.ReplaceGlobals(log)
	a.closeLog = func() error {
		restoreLogger()
		return closeLogger()
	}
	a.database, err = db.Open(filepath.Join(base, "data.db"))
	if err != nil {
		return nil, err
	}
	todoRepo := repository.NewTodoRepository(a.database)
	items, err := todoRepo.List(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load todos: %w", err)
	}
	a.todos = store.NewTodoStore(todoRepo, items, post)
	return a, nil
}

// stop cancels store work without waiting. While the UI runs, call it on the UI thread.
func (a *App) stop() {
	if a.todos != nil {
		a.todos.Close()
	}
	if a.settings != nil {
		a.settings.Close()
	}
}

// close waits for store workers before releasing their resources. Call it after the UI loop returns or setup fails.
func (a *App) close() (err error) {
	a.stop()
	if a.todos != nil {
		a.todos.Wait()
	}
	if a.settings != nil {
		a.settings.Wait()
	}
	if a.database != nil {
		err = a.database.Close()
	}
	if a.closeLog != nil {
		err = errors.Join(err, a.closeLog())
	}
	return err
}
