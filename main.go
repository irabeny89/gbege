package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/irabeny89/gosqlitex"
)

// MARK: - Workers

func handleExpiredSessions(ctx context.Context, db *gosqlitex.DbClient) {
	// run in the midnight
	t := time.Now()
	n := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if n.Before(t) {
		n = n.Add(24 * time.Hour)
	}
	diff := n.Sub(t)
	ticker := time.NewTicker(diff)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			Log.Info("Cleaning up expired sessions")
			err := DeleteExpiredSessions(db)
			if err != nil {
				Log.Error("Error cleaning up expired sessions", "err", err)
			}
		}
	}
}

func handleDeletedUsers(ctx context.Context, db *gosqlitex.DbClient) {
	// run at 1am
	t := time.Now()
	n := time.Date(t.Year(), t.Month(), t.Day(), 1, 0, 0, 0, t.Location())
	if n.Before(t) {
		n = n.Add(24 * time.Hour)
	}
	diff := n.Sub(t)
	ticker := time.NewTicker(diff)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			Log.Info("Cleaning up deleted users")
			CleanupDeletedUsers(db)
		}
	}
}

// MARK: - Main

func main() {
	// this context will be used to listen for interrupt and termination signals
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// cancel the context when the application is shutting down
	defer stop()

	db, err := gosqlitex.Open(&gosqlitex.Config{})
	if err != nil {
		Log.Error("Failed to initialize database", "err", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		Log.Error("Failed to ping database", "err", err)
		os.Exit(1)
	}
	wg := new(sync.WaitGroup)
	wg.Go(func() {
		handleExpiredSessions(sigCtx, db)
	})
	wg.Go(func() {
		handleDeletedUsers(sigCtx, db)
	})
	wg.Wait()
}
