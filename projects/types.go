package projects

import (
	"errors"
	"sync"
	"time"

	"github.com/scriptmaster/openagent/auth"
)

// Project represents a project in the system
type Project struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Domain      string    `json:"domain" db:"domain"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	IsActive    bool      `json:"is_active" db:"is_active"`
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
