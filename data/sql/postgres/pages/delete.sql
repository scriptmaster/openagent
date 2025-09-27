-- name: pages/delete
DELETE FROM ai.pages WHERE id = $1

-- name: pages/soft_delete
UPDATE ai.pages 
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1
