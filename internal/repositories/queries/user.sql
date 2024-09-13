-- name: GetUserByID :one
SELECT sqlc.embed(users), sqlc.embed(user_auth) FROM users
JOIN user_auth ON users.id = user_auth.user_id
WHERE users.id = $1;

-- name: GetUserByEmail :one
SELECT sqlc.embed(users), sqlc.embed(user_auth) FROM users 
JOIN user_auth ON users.id = user_auth.user_id
WHERE users.email = $1;

-- name: GetUserByProviderID :one
SELECT sqlc.embed(users), sqlc.embed(user_auth) FROM users 
JOIN user_auth ON users.id = user_auth.user_id
WHERE user_auth.provider_id = $1
AND user_auth.provider = $2;

-- name: GetUserBySessionToken :one
SELECT sqlc.embed(users), sqlc.embed(user_auth), sqlc.embed(user_session) FROM users 
JOIN user_auth ON users.id = user_auth.user_id
JOIN user_session ON users.id = user_session.user_id
WHERE user_session.token = $1
AND user_session.user_expired = FALSE;

-- name: CreateUser :one
INSERT INTO users (given_name, family_name, email, email_verified, avatar_url) 
VALUES (sqlc.narg(given_name), sqlc.narg(family_name), sqlc.arg(email), sqlc.arg(email_verified)::boolean, sqlc.narg(avatar_url)) RETURNING *;
 
-- name: UpdateUser :one
UPDATE users SET given_name = $2, family_name = $3 WHERE id = $1 RETURNING *;

-- name: CreateUserAuth :one
INSERT INTO user_auth (user_id, value, provider, provider_id, provider_data) 
VALUES (sqlc.arg(user_id), sqlc.arg(value), sqlc.arg(provider), sqlc.arg(provider_id), sqlc.arg(provider_data)) RETURNING *;

-- name: CreateUserSession :one
INSERT INTO user_session (user_id, token, expires_at) VALUES ($1, $2, $3) RETURNING *;

-- name: ExpireUserSession :exec
UPDATE user_session SET user_expired = TRUE WHERE user_session.id = $1;