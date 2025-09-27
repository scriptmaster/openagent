package projects

import (
	"database/sql"
	"errors"
	"log"

	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
)

// PageService defines the interface for page operations
type PageService interface {
	GetLandingPageByProjectID(projectID int) (*models.Page, error)
	GetPageByProjectIDAndSlug(projectID int, slug string) (*models.Page, error)
	CreatePage(projectID int, title, slug, htmlContent string, isLanding bool) (*models.Page, error)
	UpdatePage(id int, title, slug, htmlContent string, isLanding bool) error
	DeletePage(id int) error
	ListPagesByProjectID(projectID int) ([]*models.Page, error)
}

// pageService implements the PageService interface
type pageService struct {
	db *sql.DB
}

// NewPageService creates a new page service
func NewPageService(db *sql.DB) PageService {
	if db == nil {
		log.Fatal("database connection is required for PageService")
	}
	return &pageService{db: db}
}

// GetLandingPageByProjectID retrieves the landing page for a project
func (s *pageService) GetLandingPageByProjectID(projectID int) (*models.Page, error) {
	query, err := common.GetSQL("pages/read_landing_by_project_id")
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRow(query, projectID)
	page := &models.Page{}

	err = row.Scan(
		&page.ID, &page.ProjectID, &page.Title, &page.Slug, &page.HTMLContent,
		&page.IsLanding, &page.IsActive, &page.MetaTitle, &page.MetaDescription,
		&page.CreatedAt, &page.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no landing page found")
		}
		return nil, err
	}

	return page, nil
}

// GetPageByProjectIDAndSlug retrieves a page by project ID and slug
func (s *pageService) GetPageByProjectIDAndSlug(projectID int, slug string) (*models.Page, error) {
	query, err := common.GetSQL("pages/read_by_project_id_and_slug")
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRow(query, projectID, slug)
	page := &models.Page{}

	err = row.Scan(
		&page.ID, &page.ProjectID, &page.Title, &page.Slug, &page.HTMLContent,
		&page.IsLanding, &page.IsActive, &page.MetaTitle, &page.MetaDescription,
		&page.CreatedAt, &page.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("page not found")
		}
		return nil, err
	}

	return page, nil
}

// CreatePage creates a new page
func (s *pageService) CreatePage(projectID int, title, slug, htmlContent string, isLanding bool) (*models.Page, error) {
	query, err := common.GetSQL("pages/create")
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRow(query, projectID, title, slug, htmlContent, isLanding, "", "")
	page := &models.Page{}

	err = row.Scan(
		&page.ID, &page.ProjectID, &page.Title, &page.Slug, &page.HTMLContent,
		&page.IsLanding, &page.IsActive, &page.MetaTitle, &page.MetaDescription,
		&page.CreatedAt, &page.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return page, nil
}

// UpdatePage updates an existing page
func (s *pageService) UpdatePage(id int, title, slug, htmlContent string, isLanding bool) error {
	query, err := common.GetSQL("pages/update")
	if err != nil {
		return err
	}

	_, err = s.db.Exec(query, id, title, slug, htmlContent, isLanding, "", "")
	return err
}

// DeletePage deletes a page
func (s *pageService) DeletePage(id int) error {
	query, err := common.GetSQL("pages/delete")
	if err != nil {
		return err
	}

	_, err = s.db.Exec(query, id)
	return err
}

// ListPagesByProjectID lists all pages for a project
func (s *pageService) ListPagesByProjectID(projectID int) ([]*models.Page, error) {
	query, err := common.GetSQL("pages/list_by_project_id")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []*models.Page
	for rows.Next() {
		page := &models.Page{}
		err := rows.Scan(
			&page.ID, &page.ProjectID, &page.Title, &page.Slug, &page.HTMLContent,
			&page.IsLanding, &page.IsActive, &page.MetaTitle, &page.MetaDescription,
			&page.CreatedAt, &page.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}

	return pages, nil
}
