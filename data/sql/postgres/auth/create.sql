-- name: auth/create
INSERT INTO ai.users (email, password_hash, is_admin, created_at, last_logged_in)
VALUES ($1, $2, $3, $4, $5)
RETURNING id 