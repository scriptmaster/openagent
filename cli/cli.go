package cli

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/scriptmaster/openagent/common"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"

	// PostgreSQL driver
	_ "github.com/lib/pq"
)

// Use the shared QueryInfo from common package

var (
	db *sql.DB
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "openagent-cli",
	Short: "OpenAgent CLI tool for password management",
	Long:  `OpenAgent CLI tool provides commands for managing user passwords and other administrative tasks.`,
}

// generateHashCmd represents the generate-hash command
var generateHashCmd = &cobra.Command{
	Use:   "generate-hash [password]",
	Short: "Generate a bcrypt hash for a password",
	Long:  `Generate a bcrypt hash for the given password that can be used in the database.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		password := args[0]
		hash, err := generatePasswordHash(password)
		if err != nil {
			log.Fatalf("Error generating hash: %v", err)
		}
		fmt.Printf("Password: %s\n", password)
		fmt.Printf("Hash: %s\n", hash)
	},
}

// resetPasswordCmd represents the reset-password command
var resetPasswordCmd = &cobra.Command{
	Use:   "reset-password [email] [new-password]",
	Short: "Reset a user's password",
	Long:  `Reset a user's password by email address. The password will be hashed and stored in the database.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		newPassword := args[1]

		// Initialize database connection
		if err := initDB(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		// Generate password hash
		hash, err := generatePasswordHash(newPassword)
		if err != nil {
			log.Fatalf("Error generating hash: %v", err)
		}

		// Update password in database
		query := "UPDATE ai.users SET password_hash = $1 WHERE email = $2"
		result, err := db.Exec(query, hash, email)
		if err != nil {
			log.Fatalf("Error updating password: %v", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Fatalf("Error checking update result: %v", err)
		}

		if rowsAffected == 0 {
			log.Fatalf("No user found with email: %s", email)
		}

		fmt.Printf("Password successfully reset for user: %s\n", email)
	},
}

// listUsersCmd represents the list-users command
var listUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "List all users in the database",
	Long:  `List all users in the database with their basic information.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database connection
		if err := initDB(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		// Query users
		listQuery := "SELECT id, email, is_admin, created_at, last_logged_in FROM ai.users ORDER BY id"
		rows, err := db.Query(listQuery)
		if err != nil {
			log.Fatalf("Error querying users: %v", err)
		}
		defer rows.Close()

		fmt.Printf("%-5s %-30s %-10s %-20s %-20s\n", "ID", "Email", "Admin", "Created", "Last Login")
		fmt.Println("----------------------------------------------------------------------------------------")

		for rows.Next() {
			var id int
			var email string
			var isAdmin bool
			var createdAt, lastLoggedIn sql.NullTime

			err := rows.Scan(&id, &email, &isAdmin, &createdAt, &lastLoggedIn)
			if err != nil {
				log.Printf("Error scanning user row: %v", err)
				continue
			}

			createdStr := "N/A"
			if createdAt.Valid {
				createdStr = createdAt.Time.Format("2006-01-02 15:04")
			}

			lastLoginStr := "Never"
			if lastLoggedIn.Valid {
				lastLoginStr = lastLoggedIn.Time.Format("2006-01-02 15:04")
			}

			adminStr := "No"
			if isAdmin {
				adminStr = "Yes"
			}

			fmt.Printf("%-5d %-30s %-10s %-20s %-20s\n", id, email, adminStr, createdStr, lastLoginStr)
		}

		if err = rows.Err(); err != nil {
			log.Fatalf("Error iterating users: %v", err)
		}
	},
}

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query [query-name] [param1] [param2] ...",
	Short: "Execute a SQL query from data/sql/postgres",
	Long: `Execute a SQL query from the data/sql/postgres directory.
The first parameter is the query name (without the .sql extension).
Additional parameters are passed to the query as $1, $2, etc.

Examples:
  openagent-cli query auth/count_users
  openagent-cli query auth/get_user_by_email user@example.com
  openagent-cli query projects/get_by_domain example.com

If no query name is provided, lists all available queries.`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Load SQL queries
		if err := common.LoadNamedSQLFiles("./data/sql/postgres"); err != nil {
			log.Fatalf("Failed to load SQL queries: %v", err)
		}

		// If no arguments provided, list available queries
		if len(args) == 0 {
			listAvailableQueries()
			return
		}

		// Initialize database connection
		if err := initDB(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		queryName := args[0]
		queryParams := args[1:]

		// Convert parameters to interface{} slice
		params := make([]interface{}, len(queryParams))
		for i, param := range queryParams {
			params[i] = param
		}

		// Use shared function to execute query
		req := common.ExecuteRequest{
			QueryName: queryName,
			Params:    params,
		}

		response := common.ExecuteQuery(req, db)

		// Display results based on response
		if !response.Success {
			fmt.Printf("❌ %s\n", response.Error)
			return
		}

		// Display results
		if response.QueryType == "SELECT" {
			displaySelectResults(response)
		} else {
			displayModifyResults(response)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(generateHashCmd)
	rootCmd.AddCommand(resetPasswordCmd)
	rootCmd.AddCommand(listUsersCmd)
	rootCmd.AddCommand(queryCmd)
}

// initDB initializes the database connection
func initDB() error {
	var err error

	// Get database connection parameters from environment
	driver := common.GetEnvOrDefault("DB_DRIVER", "postgres")
	host := common.GetEnv("DB_HOST")
	port := common.GetEnv("DB_PORT")
	user := common.GetEnv("DB_USER")
	password := common.GetEnv("DB_PASSWORD")
	dbName := common.GetEnv("DB_NAME")

	if host == "" || port == "" || user == "" || dbName == "" {
		return fmt.Errorf("database connection parameters are not fully set (DB_HOST, DB_PORT, DB_USER, DB_NAME required)")
	}

	// Create connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	// Connect to database
	db, err = sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// generatePasswordHash generates a bcrypt hash for the given password
func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}
	return string(hash), nil
}

// listAvailableQueries lists all available SQL queries and their parameter counts
// displaySelectResults displays SELECT query results in table format
func displaySelectResults(response common.ExecuteResponse) {
	// Print header
	for i, col := range response.Columns {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Printf("%-15s", col)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", len(response.Columns)*17))

	// Print rows
	for _, row := range response.Rows {
		for i, val := range row {
			if i > 0 {
				fmt.Print(" | ")
			}
			var str string
			if val == nil {
				str = "NULL"
			} else {
				str = fmt.Sprintf("%v", val)
			}
			fmt.Printf("%-15s", str)
		}
		fmt.Println()
	}

	fmt.Printf("\nQuery executed successfully. %d rows returned.\n", response.RowCount)
}

// displayModifyResults displays INSERT/UPDATE/DELETE query results
func displayModifyResults(response common.ExecuteResponse) {
	fmt.Printf("Query executed successfully. %d rows affected.\n", response.RowCount)
}

func listAvailableQueries() {
	fmt.Println("Available SQL queries:")
	fmt.Println("====================")

	// Use shared function to get available queries
	queryGroups, err := common.GetAvailableQueries()
	if err != nil {
		fmt.Printf("❌ Error loading queries: %v\n", err)
		return
	}

	if len(queryGroups) == 0 {
		fmt.Println("No SQL queries found in data/sql/postgres/")
		return
	}

	// Sort parameter counts and display groups
	var paramCounts []int
	for count := range queryGroups {
		paramCounts = append(paramCounts, count)
	}
	sort.Ints(paramCounts)

	totalQueries := 0
	for _, count := range paramCounts {
		queries := queryGroups[count]
		totalQueries += len(queries)

		// Print group header
		if count == 0 {
			fmt.Printf("\n0 parameters:\n")
		} else {
			fmt.Printf("\n%d parameters:\n", count)
		}
		fmt.Println(strings.Repeat("-", 15))

		// Print queries in this group
		for _, query := range queries {
			if query.ParamDetails != "" {
				fmt.Printf("%-30s | %s\n", query.Name, query.ParamDetails)
			} else {
				fmt.Printf("%-30s\n", query.Name)
			}
		}
	}

	fmt.Printf("\nTotal: %d queries available\n", totalQueries)
	fmt.Println("\nUsage: openagent-cli query <query-name> [param1] [param2] ...")
}
