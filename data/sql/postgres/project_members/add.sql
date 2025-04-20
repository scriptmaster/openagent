-- name: project_members/add
INSERT INTO ai.project_members (project_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING 