-- name: admin_tokens/get_by_date
SELECT token, token_date FROM ai.admin_tokens WHERE token_date = $1 