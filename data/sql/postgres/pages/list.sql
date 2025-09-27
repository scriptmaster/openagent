-- name: pages/list_by_project_id
SELECT id, project_id, title, slug, html_content, is_landing, is_active, 
       meta_title, meta_description, created_at, updated_at
FROM ai.pages 
WHERE project_id = $1 AND is_active = TRUE
ORDER BY created_at DESC

-- name: pages/list_all
SELECT id, project_id, title, slug, html_content, is_landing, is_active, 
       meta_title, meta_description, created_at, updated_at
FROM ai.pages 
WHERE is_active = TRUE
ORDER BY created_at DESC
