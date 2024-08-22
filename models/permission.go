package models

import (
	"go.flare.io/auth/models/enum"
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
