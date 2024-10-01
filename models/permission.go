package models

import (
	"time"

	"goflare.io/auth/models/enum"
	"goflare.io/auth/sqlc"
)

// Permission is the permission for the resource and action.
type Permission struct {
	ID          uint64            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Resource    enum.ResourceType `json:"resource"`
	Action      enum.ActionType   `json:"action"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NewPermission creates a new Permission.
func NewPermission() *Permission {

	return &Permission{}
}

// ConvertFromSQLCPermission converts a SQLC permission to a Permission.
func (p *Permission) ConvertFromSQLCPermission(sqlcPermission any) *Permission {

	var name, description string
	var resource enum.ResourceType
	var action enum.ActionType

	switch sp := sqlcPermission.(type) {
	case *sqlc.GetPermissionByIDRow:
		name = sp.Name
		if sp.Description != nil {
			description = *sp.Description
		}
		resource = enum.ResourceType(sp.Resource)
		action = enum.ActionType(sp.Action)
	case *sqlc.Permission:
		name = sp.Name
		if sp.Description != nil {
			description = *sp.Description
		}
		resource = enum.ResourceType(sp.Resource)
		action = enum.ActionType(sp.Action)
	default:
		return nil
	}

	p.Name = name
	p.Description = description
	p.Resource = resource
	p.Action = action

	return p
}
