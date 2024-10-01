// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: permissions.sql

package sqlc

import (
	"context"
)

const createPermission = `-- name: CreatePermission :exec
INSERT INTO permissions (name, description, resource, action)
VALUES ($1, $2, $3, $4)
`

type CreatePermissionParams struct {
	Name        string       `json:"name"`
	Description *string      `json:"description"`
	Resource    ResourceType `json:"resource"`
	Action      ActionType   `json:"action"`
}

func (q *Queries) CreatePermission(ctx context.Context, arg CreatePermissionParams) error {
	_, err := q.db.Exec(ctx, createPermission,
		arg.Name,
		arg.Description,
		arg.Resource,
		arg.Action,
	)
	return err
}

const deletePermission = `-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1
`

func (q *Queries) DeletePermission(ctx context.Context, id uint64) error {
	_, err := q.db.Exec(ctx, deletePermission, id)
	return err
}

const getPermissionByID = `-- name: GetPermissionByID :one
SELECT name, description, resource, action FROM permissions WHERE id = $1
`

type GetPermissionByIDRow struct {
	Name        string       `json:"name"`
	Description *string      `json:"description"`
	Resource    ResourceType `json:"resource"`
	Action      ActionType   `json:"action"`
}

func (q *Queries) GetPermissionByID(ctx context.Context, id uint64) (*GetPermissionByIDRow, error) {
	row := q.db.QueryRow(ctx, getPermissionByID, id)
	var i GetPermissionByIDRow
	err := row.Scan(
		&i.Name,
		&i.Description,
		&i.Resource,
		&i.Action,
	)
	return &i, err
}
