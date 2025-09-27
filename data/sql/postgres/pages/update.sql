-- name: pages/update
UPDATE ai.pages 
SET title = $2, slug = $3, html_content = $4, is_landing = $5, 
    meta_title = $6, meta_description = $7, updated_at = NOW()
WHERE id = $1

-- name: pages/update_content
UPDATE ai.pages 
SET html_content = $2, updated_at = NOW()
WHERE id = $1

-- name: pages/update_landing_status
UPDATE ai.pages 
SET is_landing = $2, updated_at = NOW()
WHERE id = $1
