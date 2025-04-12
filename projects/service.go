package projects

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/scriptmaster/openagent/common"
)

// ProjectService handles project-related database operations
type ProjectService struct {
	db *sql.DB
	// Cache for projects by domain
	projectsByDomain map[string]*cachedProject
	mu               sync.RWMutex
}

type cachedProject struct {
	project       *Project
	lastRefreshed time.Time
}

// NewProjectService creates a new project service
func NewProjectService(db *sql.DB) (*ProjectService, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	// Load SQL queries
	if err := common.LoadSQLQueries("data/projects.yaml"); err != nil {
		return nil, fmt.Errorf("failed to load SQL queries: %w", err)
	}

	return &ProjectService{
		db:               db,
		projectsByDomain: make(map[string]*cachedProject),
	}, nil
}

// GetProjectByDomain retrieves a project by its domain
func (s *ProjectService) GetProjectByDomain(domain string) (*Project, error) {
	// Check cache first
	s.mu.RLock()
	cached, exists := s.projectsByDomain[domain]
	s.mu.RUnlock()

	refreshInterval := 1 * time.Hour
	if exists && !cached.lastRefreshed.Before(time.Now().Add(-refreshInterval)) {
		return cached.project, nil
	}

	// Cache miss, fetch from database
	query, err := common.GetQuery(s.db, "GetProjectByDomain")
	if err != nil {
		return nil, fmt.Errorf("failed to get query: %w", err)
	}

	var project Project
	err = s.db.QueryRow(query, domain).Scan(
		&project.ID, &project.Name, &project.Description, &project.Domain,
		&project.CreatedAt, &project.UpdatedAt, &project.Status, &project.Owner)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found for domain: %s", domain)
		}
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	// Update cache
	log.Printf("Caching project for domain: %s", domain)
	s.mu.Lock()
	s.projectsByDomain[domain] = &cachedProject{
		project:       &project,
		lastRefreshed: time.Now(),
	}
	s.mu.Unlock()

	return &project, nil
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(project *Project) error {
	query, err := common.GetQuery(s.db, "CreateProject")
	if err != nil {
		return err
	}

	_, err = s.db.Exec(query,
		project.ID, project.Name, project.Description, project.Domain,
		project.CreatedAt, project.UpdatedAt, project.Status, project.Owner)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// GetProject retrieves a project by ID
func (s *ProjectService) GetProject(id string) (*Project, error) {
	query, err := common.GetQuery(s.db, "GetProject")
	if err != nil {
		return nil, err
	}

	var project Project
	err = s.db.QueryRow(query, id).Scan(
		&project.ID, &project.Name, &project.Description, &project.Domain,
		&project.CreatedAt, &project.UpdatedAt, &project.Status, &project.Owner)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

// UpdateProject updates an existing project
func (s *ProjectService) UpdateProject(project *Project) error {
	query, err := common.GetQuery(s.db, "UpdateProject")
	if err != nil {
		return err
	}

	project.UpdatedAt = time.Now()
	_, err = s.db.Exec(query,
		project.Name, project.Description, project.Domain,
		project.UpdatedAt, project.Status, project.Owner, project.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// DeleteProject removes a project by ID
func (s *ProjectService) DeleteProject(id string) error {
	query, err := common.GetQuery(s.db, "DeleteProject")
	if err != nil {
		return err
	}

	_, err = s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// ListProjects returns all projects
func (s *ProjectService) ListProjects() ([]*Project, error) {
	query, err := common.GetQuery(s.db, "ListProjects")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.Domain,
			&project.CreatedAt, &project.UpdatedAt, &project.Status, &project.Owner)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}
	return projects, nil
}
