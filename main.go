package main

import (
	"context"
	"fmt"
	"net/http"
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
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			Log.Info("Cleaning up deleted users")
			CleanupDeletedUsers(db)
		}
	}
}

func runServer(ctx context.Context, mux *http.ServeMux) {
		p, ok := os.LookupEnv("PORT")
		if !ok {
			p = "8080"
		}
		addr := fmt.Sprintf(":%s", p)
		s := &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  10 * time.Second,
		}
		Log.Info("Server starting", "addr", addr)
		go func() {
			<-ctx.Done()
			Log.Info("Server shutting down")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.Shutdown(shutdownCtx); err != nil {
				Log.Error("Server forced to shutdown", "err", err)
			}
		}()

		if err := s.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				Log.Error("Failed to start server", "err", err)
				os.Exit(1)
			}
		}
	}

// MARK: - Main

func main() {
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := gosqlitex.Open(new(gosqlitex.Config))
	if err != nil {
		Log.Error("Failed to initialize database", "err", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		Log.Error("Failed to ping database", "err", err)
		os.Exit(1)
	}
	Log.Info("Database connected")

	migDir, ok := os.LookupEnv("MIG_DIR")
	if !ok {
		migDir = "./migrations"
	}
	sep, ok := os.LookupEnv("MIG_SEP")
	if !ok {
		sep = "_"
	}
	Log.Info("Running migrations", "dir", migDir, "sep", sep)
	err = RunMigrations(migDir, sep, db)
	if err != nil {
		Log.Error("Failed to run migrations", "err", err)
		os.Exit(1)
	}
	Log.Info("Migrations applied")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	
	wg := new(sync.WaitGroup)
	wg.Go(func() {
		runServer(sigCtx, mux)
	})
	wg.Go(func() {
		handleExpiredSessions(sigCtx, db)
	})
	wg.Go(func() {
		handleDeletedUsers(sigCtx, db)
	})
	wg.Wait()
}
