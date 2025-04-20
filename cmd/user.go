package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/server"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `Add or modify users in the system`,
}

var addUserCmd = &cobra.Command{
	Use:   "add [email] [password]",
	Short: "Add a new user",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Load environment variables from root .env
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: Error loading .env file from root: %v\n", err)
		}

		// Connect to the database
		db, err := server.InitDB()
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Create user service
		userService := auth.NewUserService(db)

		email := args[0]
		password := args[1]

		user, err := userService.CreateUser(email, password)
		if err != nil {
			fmt.Printf("Error creating user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User created successfully: %s\n", user.Email)
	},
}

var makeAdminCmd = &cobra.Command{
	Use:   "make-admin [email]",
	Short: "Make a user an admin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load environment variables from root .env
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: Error loading .env file from root: %v\n", err)
		}

		// Connect to the database
		db, err := server.InitDB()
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Create user service
		userService := auth.NewUserService(db)

		email := args[0]

		err = userService.MakeUserAdmin(email)
		if err != nil {
			fmt.Printf("Error making user admin: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User %s is now an admin\n", email)
	},
}

func init() {
	RootCmd.AddCommand(userCmd)
	userCmd.AddCommand(addUserCmd)
	userCmd.AddCommand(makeAdminCmd)
}
