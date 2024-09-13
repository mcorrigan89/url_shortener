-- name: GetBlockedDomain :one
SELECT * FROM blocked_domain WHERE domain = $1;

-- name: GetBlockedUser :one
SELECT * FROM blocked_user WHERE user_id = $1;