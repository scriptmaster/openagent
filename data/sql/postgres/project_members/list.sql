-- name: project_members/list_by_project_id
SELECT u.id, u.email, u.is_admin, u.created_at, u.last_logged_in
FROM ai.users u
JOIN ai.project_members pm ON u.id = pm.user_id
WHERE pm.project_id = $1
ORDER BY u.created_at DESC

-- name: project_members/list_by_user_id
SELECT p.id, p.name, p.description, p.domain_name, p.created_at
FROM ai.projects p
JOIN ai.project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1
ORDER BY p.created_at DESC
