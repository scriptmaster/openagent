-- name: GetUserCount
SELECT COUNT(*) FROM ai.users;

-- name: GetUserByEmail
SELECT id, email, password, is_admin, created_at, last_logged_in FROM ai.users WHERE email = $1;

-- name: InsertUser
INSERT INTO ai.users (email, password, is_admin, created_at, last_logged_in)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: UpdateUserLastLogin
UPDATE ai.users
SET last_logged_in = NOW()
WHERE id = $1;

-- name: CheckAdminExists
SELECT COUNT(*)
FROM ai.users
WHERE is_admin = true;

-- name: MakeUserAdmin
UPDATE ai.users
SET is_admin = true
WHERE id = $1;

-- name: GetSession
SELECT token, user_id, created_at, expires_at
FROM ai.sessions
WHERE token = $1;

-- name: CreateSession
INSERT INTO ai.sessions (token, user_id, created_at, expires_at)
VALUES ($1, $2, $3, $4); 