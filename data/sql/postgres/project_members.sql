-- name: GetProjectMembers
SELECT u.id, u.email, u.is_admin, u.created_at, u.last_logged_in
FROM ai.users u
JOIN ai.project_members pm ON u.id = pm.user_id
WHERE pm.project_id = $1;

-- name: AddProjectMember
INSERT INTO ai.project_members (project_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveProjectMember
DELETE FROM ai.project_members WHERE project_id = $1 AND user_id = $2; 