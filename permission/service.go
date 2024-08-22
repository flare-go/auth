package permission

import (
	"context"
	"go.flare.io/auth/models"
)

type Service interface {
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uint32) (*models.Permission, error)
	Delete(ctx context.Context, id uint32) error
}
