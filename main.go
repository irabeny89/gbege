package main

import (
	"os"
	"time"

	"github.com/irabeny89/gosqlitex"
)

// MARK: - Workers

func handleExpiredSessions(db *gosqlitex.DbClient) {
	// run in the midnight
	t := time.Now()
	n := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if n.Before(t) {
		n = n.Add(24 * time.Hour)
	}
	diff := n.Sub(t)
	ticker := time.NewTicker(diff)
	for range ticker.C {
		Log.Info("Cleaning up expired sessions")
		err := DeleteExpiredSessions(db)
		if err != nil {
			Log.Error("Error cleaning up expired sessions", "err", err)
		}
	}
}

func handleDeletedUsers(db *gosqlitex.DbClient) {
	// run at 1am
	t := time.Now()
	n := time.Date(t.Year(), t.Month(), t.Day(), 1, 0, 0, 0, t.Location())
	if n.Before(t) {
		n = n.Add(24 * time.Hour)
	}
	diff := n.Sub(t)
	ticker := time.NewTicker(diff)
	for range ticker.C {
		Log.Info("Cleaning up deleted users")
		CleanupDeletedUsers(db)
	}
}

// MARK: - Main

func main() {
	db, err := gosqlitex.Open(&gosqlitex.Config{})
	if err != nil {
		Log.Error("Failed to initialize database", "err", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		Log.Error("Failed to ping database", "err", err)
		os.Exit(1)
	}

	go handleExpiredSessions(db)
	go handleDeletedUsers(db)
}
