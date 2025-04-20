-- name: auth/update_last_login
UPDATE ai.users SET last_logged_in = $1 WHERE id = $2 