-- name: admin_tokens/get_token_by_date
SELECT token FROM ai.admin_tokens WHERE token_date = $1 