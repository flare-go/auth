// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: roles.sql

package sqlc

import (
	"context"
)

const createRole = `-- name: CreateRole :exec
INSERT INTO roles (name, description)
VALUES ($1, $2)
`

type CreateRoleParams struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (q *Queries) CreateRole(ctx context.Context, arg CreateRoleParams) error {
	_, err := q.db.Exec(ctx, createRole, arg.Name, arg.Description)
	return err
}

const deleteRole = `-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1
`

func (q *Queries) DeleteRole(ctx context.Context, id uint32) error {
	_, err := q.db.Exec(ctx, deleteRole, id)
	return err
}

const getRoleByID = `-- name: GetRoleByID :one
SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1
`

func (q *Queries) GetRoleByID(ctx context.Context, id uint32) (*Role, error) {
	row := q.db.QueryRow(ctx, getRoleByID, id)
	var i Role
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}
