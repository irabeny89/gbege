package main

import (
	"log/slog"
	"os"
)

type AppLogger struct {
	slog.Logger
}

// MARK: - Type, Const & Var

func NewAppLogger() *AppLogger {
	return &AppLogger{
		Logger: *slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}
