package server

import (
	"context" // Added context for service interfaces if needed later
	"database/sql"
	"sync"
	"time"

	// Needed if structs reference auth.User directly
	"github.com/scriptmaster/openagent/auth"
	// "github.com/scriptmaster/openagent/common" // REMOVED - Unused import
)

// --- Structs moved from database.go ---

// Project represents a project
type Project struct {
	ID          int
	Name        string
	Description string
	DomainName  string
	CreatedBy   int // User ID
	CreatedAt   time.Time
}

// ProjectDB represents a database connection configured for a project
type ProjectDB struct {
	ID               int    `db:"id"`
	ProjectID        int    `db:"project_id"`
	Name             string `db:"name"`
	Description      string `db:"description"`
	DBType           string `db:"db_type"`
	ConnectionString string `db:"connection_string"` // Encoded
	SchemaName       string `db:"schema_name"`
	IsDefault        bool   `db:"is_default"`
	CreatedAt        string `db:"created_at"` // Assuming time.Time maps ok
}

// ProjectService provides methods to work with projects
type ProjectService struct {
	db *sql.DB
}

// TableService provides methods to work with tables (Likely deprecated or refactored, consider removal if unused)
type TableService struct {
	db *sql.DB
}

// ManagedTable represents a table managed by the system within a project's database
type ManagedTable struct {
	ID          int    `db:"id"`
	ProjectID   int    `db:"project_id"`
	ProjectDBID int    `db:"project_db_id"`
	Name        string `db:"name"`
	SchemaName  string `db:"schema_name"`
	Description string `db:"description"`
	Initialized bool   `db:"initialized"`
	ReadOnly    bool   `db:"read_only"`
	CreatedAt   string `db:"created_at"`
}

// ManagedColumn represents a column within a ManagedTable
type ManagedColumn struct {
	ID             int    `db:"id"`
	ManagedTableID int    `db:"managed_table_id"`
	Name           string `db:"name"`
	DisplayName    string `db:"display_name"`
	DataType       string `db:"data_type"`
	ColumnType     string `db:"type"`
	Ordinal        int    `db:"ordinal"`
	Visible        bool   `db:"visible"`
	CreatedAt      string `db:"created_at"`
}

// DatabaseMetadataService provides methods for retrieving database metadata
type DatabaseMetadataService struct {
	db          *sql.DB
	dataService DataAccessService // Interface for data access
}

// DataAccessService defines the interface for accessing project database data
type DataAccessService interface {
	// Define methods needed by DatabaseMetadataService, e.g.:
	getConnection(ctx context.Context, projectDBID int) (*sql.DB, error)
	// Potentially others like ListSchemas, ListTables, GetColumns etc. if they live here
	// CloseConnections() // Might be needed if this service manages connections
}

// DirectDataService provides methods to interact directly with project databases
type DirectDataService struct {
	db            *sql.DB         // Connection to the main application DB
	dbConnections map[int]*sql.DB // Cache for project DB connections
	mu            sync.RWMutex
}

// TableMetadata holds combined information about a table
type TableMetadata struct {
	SchemaName     string `json:"schema_name"`
	TableName      string `json:"table_name"`
	IsManaged      bool   `json:"is_managed"`
	ManagedTableID int    `json:"managed_table_id"`
	Description    string `json:"description"`
	Initialized    bool   `json:"initialized"`
	ReadOnly       bool   `json:"read_only"`
}

// ColumnMetadata holds combined information about a column
type ColumnMetadata struct {
	ColumnName      string  `json:"column_name"`
	DataType        string  `json:"data_type"`
	IsNullable      bool    `json:"is_nullable"`
	OrdinalPosition int     `json:"ordinal_position"`
	ColumnDefault   *string `json:"column_default"`
	IsManaged       bool    `json:"is_managed"`
	ManagedColumnID int     `json:"managed_column_id"`
	DisplayName     string  `json:"display_name"`
	Visible         bool    `json:"visible"`
	SystemType      string  `json:"system_type"`
}

// SettingsService provides methods to work with settings
type SettingsService struct {
	db *sql.DB
}

// Setting represents a system setting
type Setting struct {
	ID          int    `db:"id"`
	Key         string `db:"key"`
	Value       string `db:"value"`
	Description string `db:"description"`
	Scope       string `db:"scope"`
	ScopeID     *int   `db:"scope_id"`   // Use pointer for nullable int
	UpdatedAt   string `db:"updated_at"` // Assuming time.Time maps ok
}

// PageData holds common data for HTML templates
type PageData struct {
	AppName     string
	PageTitle   string
	User        *auth.User // Use pointer from auth package
	AppVersion  string
	CurrentHost string
	Error       string
}
