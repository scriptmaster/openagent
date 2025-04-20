-- name: auth/count_admins
SELECT COUNT(*) FROM ai.users WHERE is_admin = true 