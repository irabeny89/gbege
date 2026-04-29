package main

import (
	"context"
	"time"

	"github.com/irabeny89/gbege/internal/logger"
	"github.com/irabeny89/gbege/internal/session"
	"github.com/irabeny89/gbege/internal/user"
	"github.com/irabeny89/gosqlitex"
)

func handleExpiredSessions(ctx context.Context, db *gosqlitex.DbClient) {
	// run in the midnight
	t := time.Now()
	n := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if n.Before(t) {
		n = n.Add(24 * time.Hour)
	}
	diff := n.Sub(t)
	ticker := time.NewTicker(diff)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Log.Info("Cleaning up expired sessions")
			err := session.DeleteExpiredSessions(db)
			if err != nil {
				logger.Log.Error("Error cleaning up expired sessions", "err", err)
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
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Log.Info("Cleaning up deleted users")
			err := user.CleanupDeletedUsers(db)
			if err != nil {
				logger.Log.Error("Error cleaning up deleted users", "err", err)
			}
		}
	}
}