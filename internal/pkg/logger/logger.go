package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"go-desktop-template/internal/config"
)

func New(cfg config.LogConfig) (*zap.Logger, func() error, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, nil, fmt.Errorf("parse log level: %w", err)
	}
	dir := cfg.LogDir
	if !filepath.IsAbs(dir) {
		return nil, nil, fmt.Errorf("log dir is not absolute: %s", dir)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, nil, fmt.Errorf("create log directory: %w", err)
	}
	filename := filepath.Join(dir, "app.log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, nil, err
	}
	writer := &lumberjack.Logger{
		Filename: filename,
		MaxSize:  100,
		MaxAge:   cfg.MaxAge,
		Compress: true,
	}
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	log := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoder), zapcore.AddSync(writer), level), zap.AddCaller())
	return log, func() error {
		return errors.Join(log.Sync(), writer.Close())
	}, nil
}
