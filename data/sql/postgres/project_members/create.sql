-- name: project_members/create
INSERT INTO ai.project_members (project_id, user_id) 
VALUES ($1, $2) 
ON CONFLICT (project_id, user_id) DO NOTHING
RETURNING project_id, user_id, created_at
