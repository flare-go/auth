package models

import (
	"goflare.io/auth/sqlc"
	"time"
)

type Role struct {
	ID          uint32    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewRole() *Role {
	return &Role{}
}

func (r *Role) ConvertFromSQLCRole(sqlcRole any) *Role {

	var id uint32
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
