package projects

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/scriptmaster/openagent/auth"
)

// ProjectOptions defines the structure for the JSONB options column
type ProjectOptions map[string]interface{}

// Value implements the driver.Valuer interface for ProjectOptions.
// This allows us to save the map directly to JSONB.
func (a ProjectOptions) Value() (driver.Value, error) {
	if len(a) == 0 {
		// Handle empty map explicitly to store {} instead of NULL
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for ProjectOptions.
// This allows us to read JSONB from the DB into the map.
func (a *ProjectOptions) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		// Handle the case where the DB value is nil or not []byte
		// If nil, initialize an empty map
		if value == nil {
			*a = make(ProjectOptions)
			return nil
		}
		return errors.New("type assertion to []byte failed")
	}
	// If the byte slice is nil or empty, treat as an empty JSON object
	if len(b) == 0 {
		*a = make(ProjectOptions)
		return nil
	}
	return json.Unmarshal(b, &a)
}

// Project represents a project in the system
type Project struct {
	ID          int64          `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	Domain      string         `json:"domain" db:"domain_name"`
	Options     ProjectOptions `json:"options,omitempty" db:"options"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy   int64          `json:"created_by" db:"created_by"`
	IsActive    bool           `json:"is_active" db:"is_active"`
}

// ProjectStore manages project data persistence
// Consider if this needs to be exported if only used by service.go
type ProjectStore struct {
	mu       sync.RWMutex
	projects map[string]*Project
	filePath string
}

// ProjectsPageData holds data for the projects listing page template.
type ProjectsPageData struct {
	AppName    string
	PageTitle  string
	User       auth.User
	Projects   []*Project // Changed from interface{}
	AppVersion string
}

// ProjectService defines the interface for project operations
type ProjectService interface {
	Create(project *Project) (int64, error)
	GetByID(id int64) (*Project, error)
	GetByDomain(domain string) (*Project, error)
	List() ([]*Project, error)
	Update(project *Project) error
	Delete(id int64) error
}

// ProjectRepository defines the interface for project data access
type ProjectRepository interface {
	Create(project *Project) (int64, error)
	GetByID(id int64) (*Project, error)
	GetByDomain(domain string) (*Project, error)
	List() ([]*Project, error)
	Update(project *Project) error
	Delete(id int64) error
}

// ErrProjectNotFound is returned when a project is not found
var ErrProjectNotFound = errors.New("project not found")
