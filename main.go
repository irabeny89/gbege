package main

import (
	// "net"
	// "net/http"
	"os"
	// "os/signal"
	"sync"
	// "syscall"

	_ "github.com/irabeny89/go-envy" // Load env vars
)


func main() {
	db, err := sync.OnceValues(NewDbClient)()
	if err != nil {
		Log.Error("Failed to initialize database", "err", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		Log.Error("Failed to ping database", "err", err)
		os.Exit(1)
	}

	var (
		// default host address. `:0` means random port
		// addr      = ":0"
		// // Create a channel to listen for OS signals
		// // We use a buffer of 1 so we don't miss the signal
		// sigChan   = make(chan os.Signal, 1)
	)

	Log.Info("App environment", "env", os.Getenv("APP_ENV"))
	// set address port if PORT env var exist
	// p, pExist := os.LookupEnv("PORT")
	// if pExist {
	// 	addr = ":" + p
	// }
	// if !pExist {
	// 	Log.Warn("PORT env var not found, using random port")
	// }

	// Register the signals we want to capture
	// SIGINT = Ctrl+C, SIGTERM = Standard termination (Docker/Linux)
	// signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// // Run your app logic in a separate goroutine if needed,
	// // or just wait for the signal here.
	// go func() {
	// 	ln, err := net.Listen("tcp", addr)
	// 	if err != nil {
	// 		Log.Error("Failed to start server", "err", err)
	// 		os.Exit(1)
	// 	}
	// 	Log.Info("Server is running", "addr", addr)
	// 	// Your web server or app logic goes here
	// 	if err := http.Serve(ln, nil); err != nil {
	// 		Log.Error("Server closed", "err", err)
	// 	}
	// }()

	// go CleanupDeletedUsers(db)

	// //! Block here until a signal is received
	// sig := <-sigChan
	// Log.Info("Shutting down...", "signal", sig)

	// Log.Info("Goodbye!")
}
