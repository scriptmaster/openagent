-- name: auth/get_user_by_email
SELECT id, email, password_hash, is_admin, created_at, last_logged_in FROM ai.users WHERE email = $1 