package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"go.uber.org/zap"

	"go-desktop-template/internal/config"
	"go-desktop-template/internal/db"
	"go-desktop-template/internal/desktop"
	"go-desktop-template/internal/pkg/logger"
	"go-desktop-template/internal/repository"
	"go-desktop-template/internal/service"
)

func Run(configPath string) (runErr error) {
	initialConfig, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	settings := service.NewSettings(filepath.Join(filepath.Dir(configPath), "settings.json"))
	preferences, err := settings.Load(configPath)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
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
	return desktop.Run(desktop.Options{
		Todos:       todos,
		Items:       items,
		Settings:    settings,
		Preferences: preferences,
		DataPath:    initialConfig.DataDir,
	})
}
