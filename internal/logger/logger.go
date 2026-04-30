package logger

import (
	"log/slog"
	"os"
)

var Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// SetLevel sets the log level.
func SetLevel(level slog.Level) {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
