package models

import (
	"time"

	"goflare.io/auth/internal/sqlc"
)

// User is the user for the application.
type User struct {

	// ID is the ID of the user.
	ID uint64 `json:"id"`

	// Username is the username of the user.
	Username string `json:"username"`

	// PasswordHash is the password hash of the user.
	PasswordHash string `json:"-"`

	// Email is the email of the user.
	Email string `json:"email"`

	// Phone is the phone of the user.
	Phone string `json:"phone"`

	// FirebaseUID is the Firebase UID of the user.
	FirebaseUID string `json:"firebase_uid"`

	// Provider is the provider of the user.
	Provider string `json:"provider"`

	// DisplayName is the display name of the user.
	DisplayName string `json:"display_name"`

	// PhotoURL is the photo URL of the user.
	PhotoURL string `json:"photo_url"`

	// CreatedAt is the created at time.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the updated at time.
	UpdatedAt time.Time `json:"updated_at"`

	// LastSignInAt is the last sign-in at time.
	LastSignInAt time.Time `json:"last_sign_in_at"`
}

// ConvertFromSQLCUser converts an SQLC user to a User.
func (u *User) ConvertFromSQLCUser(sqlcUser any) *User {
	type userFields struct {
		ID           uint64
		Username     string
		PasswordHash string
		Email        string
		FirebaseUID  string
		Provider     string
	}

	var fields userFields

	switch sp := sqlcUser.(type) {
	case *sqlc.FindUserByIDRow:
		fields = userFields{
			Username:     sp.Username,
			PasswordHash: sp.PasswordHash,
			Email:        sp.Email,
		}
	case *sqlc.FindUserByUsernameRow:
		fields = userFields{
			ID:           sp.ID,
			PasswordHash: sp.PasswordHash,
			Email:        sp.Email,
		}
	case *sqlc.FindUserByEmailRow:
		fields = userFields{
			ID:           sp.ID,
			Username:     sp.Username,
			PasswordHash: sp.PasswordHash,
			FirebaseUID:  *sp.FirebaseUid,
			Provider:     string(sp.Provider),
		}
	case *sqlc.ListUsersRow:
		fields = userFields{
			ID:       sp.ID,
			Username: sp.Username,
			Email:    sp.Email,
		}
	case *sqlc.FindUserByFirebaseUIDRow:
		fields = userFields{
			ID:           sp.ID,
			Username:     sp.Username,
			PasswordHash: sp.PasswordHash,
			Email:        sp.Email,
		}
	default:
		return nil
	}

	u.ID = fields.ID
	u.Username = fields.Username
	u.Email = fields.Email
	u.PasswordHash = fields.PasswordHash
	u.FirebaseUID = fields.FirebaseUID
	u.Provider = fields.Provider

	return u
}
