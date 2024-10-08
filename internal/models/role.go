package models

import (
	"time"

	"goflare.io/auth/internal/sqlc"
)

// Role is the role for the user.
type Role struct {
	// ID is the ID of the role.
	ID uint64 `json:"id"`

	// Name is the name of the role.
	Name string `json:"name"`

	// Description is the description of the role.
	Description string `json:"description"`

	// CreatedAt is the created at time.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the updated at time.
	UpdatedAt time.Time `json:"updated_at"`
}

// ConvertFromSQLCRole converts a SQLC role to a Role.
func (r *Role) ConvertFromSQLCRole(sqlcRole any) *Role {

	var id uint64
	var name, description string

	switch sp := sqlcRole.(type) {
	case *sqlc.Role:
		name = sp.Name
		if sp.Description != nil {
			description = *sp.Description
		}
	case *sqlc.ListRolesRow:
		id = sp.ID
		name = sp.Name
		if sp.Description != nil {
			description = *sp.Description
		}
	default:
		return nil
	}

	r.ID = id
	r.Name = name
	r.Description = description

	return r
}
