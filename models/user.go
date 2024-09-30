package models

import (
	"goflare.io/auth/sqlc"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	FirebaseUID  string    `json:"firebase_uid"`
	Provider     string    `json:"provider"`
	DisplayName  string    `json:"display_name"`
	PhotoURL     string    `json:"photo_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastSignInAt time.Time `json:"last_sign_in_at"`
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

	u.ID = int(id)
	u.Username = username
	u.PasswordHash = passwordHash
	u.Email = email

	return u
}
