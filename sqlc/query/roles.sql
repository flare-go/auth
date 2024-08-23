-- name: CreateRole :exec
INSERT INTO roles (name, description)
VALUES ($1, $2);

-- name: GetRoleByID :one
SELECT name, description FROM roles WHERE id = $1;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- name: ListRoles :many
SELECT id, name, description FROM roles;