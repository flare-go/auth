// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (username, password_hash, email)
VALUES ($1, $2, $3)
RETURNING id
`

type CreateUserParams struct {
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	Email        string `json:"email"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (uint64, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Username, arg.PasswordHash, arg.Email)
	var id uint64
	err := row.Scan(&id)
	return id, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id uint64) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const findUserByEmail = `-- name: FindUserByEmail :one
SELECT id, password_hash, username, created_at, updated_at  FROM users WHERE email = $1
`

type FindUserByEmailRow struct {
	ID           uint64             `json:"id"`
	PasswordHash string             `json:"passwordHash"`
	Username     string             `json:"username"`
	CreatedAt    pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt    pgtype.Timestamptz `json:"updatedAt"`
}

func (q *Queries) FindUserByEmail(ctx context.Context, email string) (*FindUserByEmailRow, error) {
	row := q.db.QueryRow(ctx, findUserByEmail, email)
	var i FindUserByEmailRow
	err := row.Scan(
		&i.ID,
		&i.PasswordHash,
		&i.Username,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const findUserByFirebaseUID = `-- name: FindUserByFirebaseUID :one
SELECT id, password_hash, username, email, created_at, updated_at  FROM users WHERE firebase_uid = $1
`

type FindUserByFirebaseUIDRow struct {
	ID           uint64             `json:"id"`
	PasswordHash string             `json:"passwordHash"`
	Username     string             `json:"username"`
	Email        string             `json:"email"`
	CreatedAt    pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt    pgtype.Timestamptz `json:"updatedAt"`
}

func (q *Queries) FindUserByFirebaseUID(ctx context.Context, firebaseUid *string) (*FindUserByFirebaseUIDRow, error) {
	row := q.db.QueryRow(ctx, findUserByFirebaseUID, firebaseUid)
	var i FindUserByFirebaseUIDRow
	err := row.Scan(
		&i.ID,
		&i.PasswordHash,
		&i.Username,
		&i.Email,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const findUserByID = `-- name: FindUserByID :one
SELECT username, password_hash, email, created_at, updated_at  FROM users WHERE id = $1
`

type FindUserByIDRow struct {
	Username     string             `json:"username"`
	PasswordHash string             `json:"passwordHash"`
	Email        string             `json:"email"`
	CreatedAt    pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt    pgtype.Timestamptz `json:"updatedAt"`
}

func (q *Queries) FindUserByID(ctx context.Context, id uint64) (*FindUserByIDRow, error) {
	row := q.db.QueryRow(ctx, findUserByID, id)
	var i FindUserByIDRow
	err := row.Scan(
		&i.Username,
		&i.PasswordHash,
		&i.Email,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const findUserByUsername = `-- name: FindUserByUsername :one
SELECT id, password_hash, email, created_at, updated_at  FROM users WHERE username = $1
`

type FindUserByUsernameRow struct {
	ID           uint64             `json:"id"`
	PasswordHash string             `json:"passwordHash"`
	Email        string             `json:"email"`
	CreatedAt    pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt    pgtype.Timestamptz `json:"updatedAt"`
}

func (q *Queries) FindUserByUsername(ctx context.Context, username string) (*FindUserByUsernameRow, error) {
	row := q.db.QueryRow(ctx, findUserByUsername, username)
	var i FindUserByUsernameRow
	err := row.Scan(
		&i.ID,
		&i.PasswordHash,
		&i.Email,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, username, email, created_at, updated_at FROM users
`

type ListUsersRow struct {
	ID        uint64             `json:"id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	CreatedAt pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt pgtype.Timestamptz `json:"updatedAt"`
}

func (q *Queries) ListUsers(ctx context.Context) ([]*ListUsersRow, error) {
	rows, err := q.db.Query(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*ListUsersRow{}
	for rows.Next() {
		var i ListUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUserEmail = `-- name: UpdateUserEmail :exec
UPDATE users
SET email = $2, updated_at = NOW()
WHERE id = $1
`

type UpdateUserEmailParams struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
}

func (q *Queries) UpdateUserEmail(ctx context.Context, arg UpdateUserEmailParams) error {
	_, err := q.db.Exec(ctx, updateUserEmail, arg.ID, arg.Email)
	return err
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1
`

type UpdateUserPasswordParams struct {
	ID           uint64 `json:"id"`
	PasswordHash string `json:"passwordHash"`
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.Exec(ctx, updateUserPassword, arg.ID, arg.PasswordHash)
	return err
}

const updateUsername = `-- name: UpdateUsername :exec
UPDATE users
SET username = $2, updated_at = NOW()
WHERE id = $1
`

type UpdateUsernameParams struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
}

func (q *Queries) UpdateUsername(ctx context.Context, arg UpdateUsernameParams) error {
	_, err := q.db.Exec(ctx, updateUsername, arg.ID, arg.Username)
	return err
}
