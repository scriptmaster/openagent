package projects

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Project represents a project in the system
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Status      string    `json:"status"`
	Owner       string    `json:"owner"`
	Members     []string  `json:"members"`
}

// ProjectStore manages project data persistence
type ProjectStore struct {
	mu       sync.RWMutex
	projects map[string]*Project
	filePath string
}

// NewProjectStore creates a new project store
func NewProjectStore(dataDir string) (*ProjectStore, error) {
	store := &ProjectStore{
		projects: make(map[string]*Project),
		filePath: filepath.Join(dataDir, "projects.json"),
	}

	if err := store.load(); err != nil {
		return nil, fmt.Errorf("failed to load projects: %w", err)
	}

	return store, nil
}

// load reads projects from the JSON file
func (s *ProjectStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.projects)
}

// save writes projects to the JSON file
func (s *ProjectStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.projects, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// CreateProject creates a new project
func (s *ProjectStore) CreateProject(project *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[project.ID]; exists {
		return fmt.Errorf("project with ID %s already exists", project.ID)
	}

	s.projects[project.ID] = project
	return s.save()
}

// GetProject retrieves a project by ID
func (s *ProjectStore) GetProject(id string) (*Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, exists := s.projects[id]
	if !exists {
		return nil, fmt.Errorf("project with ID %s not found", id)
	}

	return project, nil
}

// UpdateProject updates an existing project
func (s *ProjectStore) UpdateProject(project *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[project.ID]; !exists {
		return fmt.Errorf("project with ID %s not found", project.ID)
	}

	s.projects[project.ID] = project
	return s.save()
}

// DeleteProject removes a project by ID
func (s *ProjectStore) DeleteProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[id]; !exists {
		return fmt.Errorf("project with ID %s not found", id)
	}

	delete(s.projects, id)
	return s.save()
}

// ListProjects returns all projects
func (s *ProjectStore) ListProjects() []*Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	projects := make([]*Project, 0, len(s.projects))
	for _, project := range s.projects {
		projects = append(projects, project)
	}

	return projects
}
