package models

import (
	"goflare.io/auth/sqlc"
	"time"
)

type User struct {
	ID           uint32    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewUser() *User {
	return &User{}
}

func (u *User) ConvertFromSQLCUser(sqlcUser any) *User {

	var username, passwordHash, email string
	var id uint32

	switch sp := sqlcUser.(type) {
	case *sqlc.GetUserByIDRow:
		username = sp.Username
		passwordHash = sp.PasswordHash
		email = sp.Email
	case *sqlc.GetUserByUsernameRow:
		id = sp.ID
		passwordHash = sp.PasswordHash
		email = sp.Email
	case *sqlc.GetUserByEmailRow:
		id = sp.ID
		passwordHash = sp.PasswordHash
		username = sp.Username
	case *sqlc.ListUsersRow:
		id = sp.ID
		username = sp.Username
		email = sp.Email
	default:
		return nil
	}

	u.ID = id
	u.Username = username
	u.PasswordHash = passwordHash
	u.Email = email

	return u
}
