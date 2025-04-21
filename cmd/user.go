package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

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
	Short: "Add a new user with password",
	Long:  `Adds a new user to the default database with the specified email and password.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		password := args[1]

		userService := initializeUserServiceForCmd()

		// Hash the password
		hash, err := auth.GeneratePasswordHash(password)
		if err != nil {
			fmt.Printf("Error hashing password: %v\n", err)
			os.Exit(1)
		}

		// Use context.Background for CLI operations
		ctx := context.Background()

		// Check if user already exists
		_, err = userService.GetUserByEmail(ctx, email)
		if err == nil {
			fmt.Printf("Error: User with email %s already exists\n", email)
			os.Exit(1)
		} else if !strings.Contains(err.Error(), "user not found") {
			fmt.Printf("Error checking existing user: %v\n", err)
			os.Exit(1)
		}

		// Create the user (CreateUser now handles admin logic)
		newUser, err := userService.CreateUser(ctx, email)
		if err != nil {
			fmt.Printf("Error creating user: %v\n", err)
			os.Exit(1)
		}

		// Update the password hash (requires a new service method)
		if err := updatePasswordHash(userService, ctx, newUser.ID, hash); err != nil {
			fmt.Printf("Error setting password for user %s: %v\n", email, err)
			os.Exit(1)
		}

		fmt.Printf("User created successfully: %s\n", newUser.Email)
	},
}

var makeAdminCmd = &cobra.Command{
	Use:   "make-admin [email]",
	Short: "Make a user an admin",
	Long:  `Grants administrative privileges to an existing user specified by email.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		userService := initializeUserServiceForCmd()

		// Use context.Background for CLI operations
		ctx := context.Background()

		// Find user by email
		user, err := userService.GetUserByEmail(ctx, email)
		if err != nil {
			fmt.Printf("Error finding user %s: %v\n", email, err)
			os.Exit(1)
		}

		// Make user admin
		err = userService.MakeUserAdmin(ctx, user.ID)
		if err != nil {
			fmt.Printf("Error making user %s admin: %v\n", email, err)
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

func initializeUserServiceForCmd() auth.UserServicer {
	// Load environment variables from root .env
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Error loading .env file from root: %v\n", err)
	}

	// Initialize the database (MUST be done before initializing services)
	db, err := server.InitDB()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	// Note: We don't defer db.Close() here as the command finishes quickly.

	// Initialize dependent services (similar to server/routes.go)
	pdbService := server.NewProjectDBService(db)
	dataService := server.NewDirectDataService(db)
	userService := server.NewUserService(db, pdbService, dataService)

	return userService // userService is *server.UserService which implements auth.UserServicer
}

func updatePasswordHash(userService auth.UserServicer, ctx context.Context, userID int, hash string) error {
	// TODO: Implement this method in the server.UserService
	// It should update the password_hash field for the given userID.
	log.Printf("TODO: Implement password update logic for user %d in UserService", userID)
	// This requires adding an UpdatePassword method to the UserServicer interface
	// and implementing it in server.UserService
	// For now, return an error
	return errors.New("updatePasswordHash not implemented")
}
