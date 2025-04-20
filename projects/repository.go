package projects

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type projectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *Project) (int64, error) {
	query := `
		INSERT INTO projects (name, description, domain, created_by, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	var id int64
	err := r.db.QueryRow(
		query,
		project.Name,
		project.Description,
		project.Domain,
		project.CreatedBy,
		project.IsActive,
		project.CreatedAt,
		project.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	project.ID = id
	return id, nil
}

func (r *projectRepository) GetByID(id int64) (*Project, error) {
	project := &Project{}
	query := `SELECT * FROM projects WHERE id = $1`
	err := r.db.Get(project, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return project, err
}

func (r *projectRepository) GetByDomain(domain string) (*Project, error) {
	project := &Project{}
	query := `SELECT * FROM projects WHERE domain = $1`
	err := r.db.Get(project, query, domain)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return project, err
}

func (r *projectRepository) List() ([]*Project, error) {
	var projects []*Project
	query := `SELECT * FROM projects ORDER BY created_at DESC`
	err := r.db.Select(&projects, query)
	return projects, err
}

func (r *projectRepository) Update(project *Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, domain = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`

	project.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		query,
		project.Name,
		project.Description,
		project.Domain,
		project.IsActive,
		project.UpdatedAt,
		project.ID,
	)
	return err
}

func (r *projectRepository) Delete(id int64) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
