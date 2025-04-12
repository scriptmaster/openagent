package server

import (
	"context" // Added context for service interfaces if needed later
	"database/sql"
	"sync"
	"time"
	// Needed if structs reference auth.User directly
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
	ID               int
	ProjectID        int
	Name             string
	Description      string
	DBType           string // e.g., "postgresql"
	ConnectionString string // Base64 encoded
	SchemaName       string
	IsDefault        bool
	CreatedAt        time.Time
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
	ID          int
	ProjectID   int
	ProjectDBID int
	Name        string
	SchemaName  string
	Description string
	Initialized bool
	ReadOnly    bool
	CreatedAt   time.Time
	Columns     []ManagedColumn `json:"columns,omitempty"` // Include columns when needed
}

// ManagedColumn represents a column within a ManagedTable
type ManagedColumn struct {
	ID             int
	ManagedTableID int
	Name           string
	DisplayName    string
	DataType       string // Original DB data type
	ColumnType     string // System type (text, number, date, boolean)
	Ordinal        int    // Added field for ordinal position
	IsPrimaryKey   bool
	IsNullable     bool
	DefaultValue   sql.NullString
	Visible        bool
	IsGenerated    bool // e.g., SERIAL, auto-increment
	CreatedAt      time.Time
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
	SchemaName     string           `json:"schema_name"`
	TableName      string           `json:"table_name"`
	Description    string           `json:"description"` // From managed_tables if managed
	IsManaged      bool             `json:"is_managed"`
	ManagedTableID int              `json:"managed_table_id,omitempty"`
	Columns        []ColumnMetadata `json:"columns"`
	Initialized    bool             `json:"initialized"` // From managed_tables
	ReadOnly       bool             `json:"read_only"`   // From managed_tables
}

// ColumnMetadata holds combined information about a column
type ColumnMetadata struct {
	ColumnName      string `json:"column_name"`
	DataType        string `json:"data_type"`
	IsNullable      string `json:"is_nullable"`             // Raw value from DB ('YES'/'NO')
	DefaultValue    string `json:"default_value,omitempty"` // Raw value from DB
	OrdinalPos      int    `json:"ordinal_pos"`             // Added field for ordinal position
	IsPrimaryKey    bool   `json:"is_primary_key"`
	IsManaged       bool   `json:"is_managed"`
	ManagedColumnID int    `json:"managed_column_id,omitempty"`
	Visible         bool   `json:"visible"`      // From managed_columns
	DisplayName     string `json:"display_name"` // From managed_columns
	SystemType      string `json:"system_type"`  // From managed_columns (text, number, etc.)
}

// SettingsService provides methods to work with settings
type SettingsService struct {
	db *sql.DB
}

// Setting represents a system setting
type Setting struct {
	ID          int
	Key         string
	Value       string
	Description string
	Scope       string // "system", "project", or "user"
	ScopeID     int    // ID of the project or user (0 for system)
	UpdatedAt   time.Time
}

// ProjectDBService provides methods to work with project database connections
type ProjectDBService struct {
	db *sql.DB
}
