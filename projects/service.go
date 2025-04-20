package projects

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/scriptmaster/openagent/common"
)

// projectService implements the ProjectService interface
type projectService struct {
	db *sql.DB
	// Cache for projects by domain
	projectsByDomain map[string]*cachedProject
	mu               sync.RWMutex
	repo             ProjectRepository
}

type cachedProject struct {
	project       *Project
	lastRefreshed time.Time
}

// NewProjectService creates a new project service
func NewProjectService(db *sql.DB, repo ProjectRepository) ProjectService {
	if db == nil {
		log.Fatal("database connection is required")
	}

	// Load SQL queries
	if err := common.LoadSQLQueries("data/projects.yaml"); err != nil {
		log.Fatalf("Failed to load SQL queries: %v", err)
	}

	return &projectService{
		db:               db,
		projectsByDomain: make(map[string]*cachedProject),
		repo:             repo,
	}
}

// Create implements ProjectService.Create
func (s *projectService) Create(project *Project) (int64, error) {
	if project.Name == "" {
		return 0, errors.New("project name is required")
	}
	if project.Domain == "" {
		return 0, errors.New("project domain is required")
	}

	// Check if domain is already taken
	existing, err := s.repo.GetByDomain(project.Domain)
	if err != nil && !errors.Is(err, ErrProjectNotFound) {
		return 0, fmt.Errorf("failed to check domain: %v", err)
	}
	if existing != nil {
		return 0, errors.New("domain is already in use")
	}

	return s.repo.Create(project)
}

// GetByID implements ProjectService.GetByID
func (s *projectService) GetByID(id int64) (*Project, error) {
	return s.repo.GetByID(id)
}

// GetByDomain implements ProjectService.GetByDomain
func (s *projectService) GetByDomain(domain string) (*Project, error) {
	// Check cache first
	s.mu.RLock()
	cached, exists := s.projectsByDomain[domain]
	s.mu.RUnlock()

	refreshInterval := 1 * time.Hour
	if exists && !cached.lastRefreshed.Before(time.Now().Add(-refreshInterval)) {
		return cached.project, nil
	}

	// Cache miss, fetch from database
	project, err := s.repo.GetByDomain(domain)
	if err != nil {
		return nil, err
	}

	// Update cache
	log.Printf("Caching project for domain: %s", domain)
	s.mu.Lock()
	s.projectsByDomain[domain] = &cachedProject{
		project:       project,
		lastRefreshed: time.Now(),
	}
	s.mu.Unlock()

	return project, nil
}

// List implements ProjectService.List
func (s *projectService) List() ([]*Project, error) {
	return s.repo.List()
}

// Update implements ProjectService.Update
func (s *projectService) Update(project *Project) error {
	if project.Name == "" {
		return errors.New("project name is required")
	}
	if project.Domain == "" {
		return errors.New("project domain is required")
	}

	// Check if domain is already taken by another project
	existing, err := s.repo.GetByDomain(project.Domain)
	if err != nil && !errors.Is(err, ErrProjectNotFound) {
		return fmt.Errorf("failed to check domain: %v", err)
	}
	if existing != nil && existing.ID != project.ID {
		return errors.New("domain is already in use by another project")
	}

	return s.repo.Update(project)
}

// Delete implements ProjectService.Delete
func (s *projectService) Delete(id int64) error {
	return s.repo.Delete(id)
}
