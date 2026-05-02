package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/irabeny89/gbege/internal/api"
	"github.com/irabeny89/gbege/internal/auth"
	"github.com/irabeny89/gbege/internal/logger"
	"github.com/irabeny89/gbege/internal/middleware"

	"github.com/irabeny89/gosqlitex"
)

func setupHandlers(db *gosqlitex.DbClient) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		api.Success(w, http.StatusOK, "Server running", nil)
	})
	// MARK: - Auth routes
	h := auth.NewHandler(db)
	mux.Handle("POST /auth/login", middleware.Tracing(http.HandlerFunc(h.HandleLogin)))
	mux.Handle("POST /auth/signup", middleware.Tracing(http.HandlerFunc(h.HandleSignUp)))
	mux.Handle("POST /auth/logout", middleware.Tracing(http.HandlerFunc(h.HandleLogout)))
	mux.Handle("GET /auth/me", middleware.Tracing(http.HandlerFunc(h.HandleMe)))

	return mux
}

func runServer(ctx context.Context, db *gosqlitex.DbClient) {
	p, ok := os.LookupEnv("PORT")
	if !ok {
		p = "8080"
	}
	addr := ":" + p
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
