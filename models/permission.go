package models

import (
	"go.flare.io/auth/models/enum"
	"go.flare.io/auth/sqlc"
	"time"
)

type Permission struct {
	ID          uint32            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Resource    enum.ResourceType `json:"resource"`
	Action      enum.ActionType   `json:"action"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func NewPermission() *Permission {

	return &Permission{}
}

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
