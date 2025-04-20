package projects

import (
	"database/sql"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/scriptmaster/openagent/common"
)

type projectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *Project) (int64, error) {
	if project.Options == nil {
		project.Options = make(ProjectOptions)
	}

	query, err := common.GetSQL("projects/create")
	if err != nil {
		return 0, err
	}

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	var id int64
	err = r.db.QueryRow(
		query,
		project.Name,
		project.Description,
		project.Domain,
		project.CreatedBy,
		project.IsActive,
		project.CreatedAt,
		project.UpdatedAt,
		project.Options,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	project.ID = id
	return id, nil
}

func (r *projectRepository) GetByID(id int64) (*Project, error) {
	project := &Project{}
	query, err := common.GetSQL("projects/get_by_id")
	if err != nil {
		return nil, err
	}
	err = r.db.Get(project, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrProjectNotFound
	}
	if project.Options == nil {
		project.Options = make(ProjectOptions)
	}
	return project, err
}

func (r *projectRepository) GetByDomain(domain string) (*Project, error) {
	project := &Project{}
	query, err := common.GetSQL("projects/get_by_domain")
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRowx(query, domain)
	err = row.StructScan(project)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProjectNotFound
		}
		log.Printf("Error scanning project struct for domain '%s': %v", domain, err)
		return nil, err
	}

	if project.Options == nil {
		project.Options = make(ProjectOptions)
	}
	return project, nil
}

func (r *projectRepository) List() ([]*Project, error) {
	var projects []*Project
	query, err := common.GetSQL("projects/list")
	if err != nil {
		return nil, err
	}
	err = r.db.Select(&projects, query)
	for _, p := range projects {
		if p.Options == nil {
			p.Options = make(ProjectOptions)
		}
	}
	return projects, err
}

func (r *projectRepository) Update(project *Project) error {
	if project.Options == nil {
		project.Options = make(ProjectOptions)
	}

	query, err := common.GetSQL("projects/update")
	if err != nil {
		return err
	}

	project.UpdatedAt = time.Now()
	_, err = r.db.Exec(
		query,
		project.Name,
		project.Description,
		project.Domain,
		project.IsActive,
		project.UpdatedAt,
		project.Options,
		project.ID,
	)
	return err
}

func (r *projectRepository) Delete(id int64) error {
	query, err := common.GetSQL("projects/delete")
	if err != nil {
		return err
	}
	_, err = r.db.Exec(query, id)
	return err
}
