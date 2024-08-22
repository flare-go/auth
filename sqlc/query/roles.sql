-- name: CreateRole :exec
INSERT INTO roles (name, description)
VALUES ($1, $2);

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;