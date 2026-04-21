package main

import (
	"log/slog"
	"os"
	"sync"
)

var (
	instance *slog.Logger
	once     sync.Once
)

// GetLogger returns a singleton logger instance.
func GetLogger() *slog.Logger {
	once.Do(func() {
		instance = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	})
	return instance
}
