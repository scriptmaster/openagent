-- name: project_members/delete
DELETE FROM ai.project_members 
WHERE project_id = $1 AND user_id = $2

-- name: project_members/delete_by_project
DELETE FROM ai.project_members 
WHERE project_id = $1

-- name: project_members/delete_by_user
DELETE FROM ai.project_members 
WHERE user_id = $1
