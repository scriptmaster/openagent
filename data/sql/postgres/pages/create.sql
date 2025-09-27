-- name: pages/create
INSERT INTO ai.pages (project_id, title, slug, html_content, is_landing, is_active, meta_title, meta_description)
VALUES ($1, $2, $3, $4, $5, TRUE, $6, $7)
RETURNING id, project_id, title, slug, html_content, is_landing, is_active, 
          meta_title, meta_description, created_at, updated_at
