-- name: pages/read_by_id
SELECT id, project_id, title, slug, html_content, is_landing, is_active, 
       meta_title, meta_description, created_at, updated_at
FROM ai.pages 
WHERE id = $1 AND is_active = TRUE

-- name: pages/read_by_project_id_and_slug
SELECT id, project_id, title, slug, html_content, is_landing, is_active, 
       meta_title, meta_description, created_at, updated_at
FROM ai.pages 
WHERE project_id = $1 AND slug = $2 AND is_active = TRUE

-- name: pages/read_landing_by_project_id
SELECT id, project_id, title, slug, html_content, is_landing, is_active, 
       meta_title, meta_description, created_at, updated_at
FROM ai.pages 
WHERE project_id = $1 AND is_landing = TRUE AND is_active = TRUE
ORDER BY created_at DESC
LIMIT 1
