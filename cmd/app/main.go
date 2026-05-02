package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/irabeny89/gbege/internal/logger"
)

func main() {
	env, ok := os.LookupEnv("APP_ENV")
	if !ok {
		env = "development"
	}
	if env == "development" {
		logger.SetLevel(slog.LevelDebug)
	} else {
		logger.SetLevel(slog.LevelInfo)
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := dbConn()
	if err != nil {
		logger.Log.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Log.Info("Database connected")

	if err := handleMigrations(sigCtx, db); err != nil {
		logger.Log.Error("Failed to run migrations", "err", err)
		os.Exit(1)
	}

	wg := new(sync.WaitGroup)
	wg.Go(func() {
		runServer(sigCtx, db)
	})
	wg.Go(func() {
		handleExpiredSessions(sigCtx, db)
	})
	wg.Go(func() {
		handleDeletedUsers(sigCtx, db)
	})
	wg.Wait()
}
