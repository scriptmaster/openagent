-- name: project_members/read_by_project_and_user
SELECT pm.project_id, pm.user_id, pm.created_at
FROM ai.project_members pm
WHERE pm.project_id = $1 AND pm.user_id = $2

-- name: project_members/read_by_project_id
SELECT pm.project_id, pm.user_id, pm.created_at
FROM ai.project_members pm
WHERE pm.project_id = $1
