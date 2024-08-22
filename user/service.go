package user

import (
	"context"
	"go.flare.io/auth/models"
)

type Service interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint32) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	AssignRoleToUserWithTx(ctx context.Context, userID, roleID uint32) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error
	GetUserRoles(ctx context.Context, userID uint32) ([]models.Role, error)
}
