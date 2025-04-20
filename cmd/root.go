package cmd

import (
	"github.com/spf13/cobra"
)

// ServerStartFunc holds the function to start the server, will be set from main.go
var ServerStartFunc func()

// RootCmd is the root command for OpenAgent CLI
var RootCmd = &cobra.Command{
	Use:   "openagent",
	Short: "OpenAgent is a powerful agent-based automation platform",
	Long: `OpenAgent is a platform for creating and managing automated agents
that can perform various tasks and workflows.`,
	// When no subcommand is specified, run the server
	RunE: func(cmd *cobra.Command, args []string) error {
		if ServerStartFunc != nil {
			ServerStartFunc()
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}
