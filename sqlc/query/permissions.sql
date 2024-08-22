-- name: CreatePermission :exec
INSERT INTO permissions (name, description, resource, action)
VALUES ($1, $2, $3, $4);

-- name: GetPermissionByID :one
SELECT name, description, resource, action FROM permissions WHERE id = $1;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;