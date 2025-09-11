-- name: auth/check_admin_exists
SELECT EXISTS(SELECT 1 FROM ai.users WHERE is_admin = true)

-- name: auth/count_admins
SELECT COUNT(*) FROM ai.users WHERE is_admin = true

-- name: auth/count_all
SELECT COUNT(*) FROM ai.users

-- name: auth/count_users
SELECT COUNT(*) FROM ai.users

-- name: auth/create
INSERT INTO ai.users (email, password_hash, is_admin, created_at, last_logged_in)
VALUES ($1, $2, $3, $4, $5)
RETURNING id

-- name: auth/get_user_by_email
SELECT id, email, password_hash, is_admin, created_at, last_logged_in FROM ai.users WHERE email = $1

-- name: auth/make_admin
UPDATE ai.users SET is_admin = true WHERE id = $1

-- name: auth/update_last_login
UPDATE ai.users SET last_logged_in = NOW() WHERE id = $1

-- name: auth/update_password_hash
UPDATE ai.users SET password_hash = $1 WHERE id = $2

-- name: auth/verify_password
SELECT id, email, password_hash, is_admin, created_at, last_logged_in FROM ai.users WHERE email = $1
