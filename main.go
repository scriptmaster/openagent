package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scriptmaster/openagent/cmd"
	"github.com/scriptmaster/openagent/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the OpenAgent server",
	Long:  `Start the OpenAgent server and listen for connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	// Set the server start function as the default action
	cmd.ServerStartFunc = startServer

	// Add server command for explicit usage
	cmd.RootCmd.AddCommand(serverCmd)
}

func main() {
	// Use command structure with default command
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startServer() {
	// Set up signal catching
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Outer loop for restarting the server
	for {
		log.Println("--- Server Starting ---")
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
