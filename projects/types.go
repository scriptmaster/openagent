package projects

import (
	"sync"
	"time"

	"github.com/scriptmaster/openagent/auth"
)

// Project represents a project in the system
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Status      string    `json:"status"`
	Owner       string    `json:"owner"`   // Consider using User ID or email
	Members     []string  `json:"members"` // Consider using User IDs or emails
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
