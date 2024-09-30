-- name: CreateUser :one
INSERT INTO users (username, password_hash, email)
VALUES ($1, $2, $3)
RETURNING id;

-- name: FindUserByID :one
SELECT username, password_hash, email, created_at, updated_at  FROM users WHERE id = $1;

-- name: FindUserByUsername :one
SELECT id, password_hash, email, created_at, updated_at  FROM users WHERE username = $1;

-- name: FindUserByEmail :one
SELECT id, password_hash, username, created_at, updated_at  FROM users WHERE email = $1;

-- name: FindUserByFirebaseUID :one
SELECT id, password_hash, username, email, created_at, updated_at  FROM users WHERE firebase_uid = $1;

-- name: UpdateUsername :exec
UPDATE users
SET username = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserEmail :exec
UPDATE users
SET email = $2, updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, email, created_at, updated_at FROM users;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;