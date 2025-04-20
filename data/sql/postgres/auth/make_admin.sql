-- name: auth/make_admin
UPDATE ai.users SET is_admin = true WHERE id = $1 