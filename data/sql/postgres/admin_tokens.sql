-- name: GetAdminTokenByDate
SELECT token, token_date FROM ai.admin_tokens WHERE token_date = $1;

-- name: InsertAdminToken
INSERT INTO ai.admin_tokens (token, token_date) VALUES ($1, $2);

-- name: GetAdminTokenForDateCheck
SELECT token FROM ai.admin_tokens WHERE token_date = $1; 