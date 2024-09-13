-- name: GetLinkByID :one
SELECT * FROM link_redirect WHERE id = $1;

-- name: GetLinkByShortenedURL :one
SELECT * FROM link_redirect WHERE shortened_url = $1;

-- name: GetLinksByUserID :many
SELECT * FROM link_redirect WHERE created_by = $1 OR updated_by = $1;

-- name: CreateLink :one
INSERT INTO link_redirect (link_url, shortened_url, created_by, updated_by) 
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateLink :one
UPDATE link_redirect SET 
link_url = COALESCE(sqlc.narg(link_url), link_url), 
active = COALESCE(sqlc.narg(active), active),  
updated_by = sqlc.arg(updated_by), 
updated_at = now(), 
version = version + 1 
WHERE id = sqlc.arg(id) RETURNING *;

-- name: CreateLinkHistory :one
INSERT INTO link_redirect_history (link_id, link_url, active, quarantined, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;