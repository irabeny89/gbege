package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	client, _ := NewDbClient()
	logger := NewAppLogger()

	addr := ":8080" // default host address
	// set address port if PORT env var exist
	p, pExist := os.LookupEnv("PORT")
	if pExist {
		addr = ":" + p
	}
	if !pExist {
		logger.Warn("PORT env var not found", "PORT", p)
	}

	// Create a channel to listen for OS signals
	// We use a buffer of 1 so we don't miss the signal
	sigChan := make(chan os.Signal, 1)

	// Register the signals we want to capture
	// SIGINT = Ctrl+C, SIGTERM = Standard termination (Docker/Linux)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run your app logic in a separate goroutine if needed,
	// or just wait for the signal here.
	go func() {
		logger.Info("Server is running...")
		// Your web server or app logic goes here
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("Server closed", "err", err)
		}
	}()

	go CleanupDeletedUsers(client, logger)

	//! Block here until a signal is received
	sig := <-sigChan
	logger.Info("Shutting down...", "signal", sig)

	// Trigger your graceful shutdown with the timeout logic
	if err := client.Close(); err != nil {
		logger.Error("Error during shutdown", "err", err)
	}

	logger.Info("Goodbye!")
}
