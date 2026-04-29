package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/irabeny89/gbege/internal/auth"
	"github.com/irabeny89/gbege/internal/logger"

	"github.com/irabeny89/gosqlitex"
)

func setupHandlers(db *gosqlitex.DbClient) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleLogin(db, w, r)
	})
	mux.HandleFunc("/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleSignUp(db, w, r)
	})
	mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleLogout(db, w, r)
	})
	mux.HandleFunc("/auth/me", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleMe(db, w, r)
	})

	return mux
}

func runServer(ctx context.Context, db *gosqlitex.DbClient) {
	p, ok := os.LookupEnv("PORT")
	if !ok {
		p = "8080"
	}
	addr := fmt.Sprintf(":%s", p)
	s := &http.Server{
		Addr:         addr,
		Handler:      setupHandlers(db),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		logger.Log.Info("Server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("Server forced to shutdown", "err", err)
		}
	}()

	logger.Log.Info("Server starting", "addr", addr)
	if err := s.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			logger.Log.Error("Failed to start server", "err", err)
			os.Exit(1)
		}
	}
}
