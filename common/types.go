package common

// ProjectDB represents the data for a project's database connection
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

// ProjectDBService defines methods for managing project database connections
type ProjectDBService interface {
	GetProjectDBs(projectID int) ([]ProjectDB, error)
	GetProjectDB(id int) (ProjectDB, error)
	CreateProjectDB(projectID int, name, description, dbType, connectionString, schemaName string, isDefault bool) (ProjectDB, error)
	UpdateProjectDB(id int, name, description, dbType, connectionString, schemaName string, isDefault bool) error
	DeleteProjectDB(id int) error
	TestConnection(projectDB ProjectDB) error
	DecodeConnectionString(encoded string) (string, error)
}

// Add other common types/interfaces here...
