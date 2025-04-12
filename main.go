package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scriptmaster/openagent/server"
)

func main() {
	// Set up signal catching
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Outer loop for restarting the server
	for {
		log.Println("Attempting to start server...")
		errCh := make(chan error, 1)
		go func() {
			errCh <- server.StartServer()
		}()

		// Wait for termination signal or server error
		select {
		case err := <-errCh:
			// Server exited or encountered an error
			if err != nil {
				log.Printf("Server exited with error: %v", err)
			} else {
				log.Println("Server exited unexpectedly without error.")
			}
			log.Println("Attempting restart in 10 seconds...")
			time.Sleep(10 * time.Second)
			// Continue to the next iteration of the for loop to restart

		case sig := <-signals:
			// Received termination signal
			appVersion := os.Getenv("APP_VERSION")
			if appVersion == "" {
				appVersion = "1.0.0.0" // Default if not set
			}

			log.Printf("\n📡 Received signal %v. Bye bye! OpenAgent version %s shutting down...\n", sig, appVersion)

			// Exit gracefully - this breaks the restart loop
			os.Exit(0)
		}
	}
}
