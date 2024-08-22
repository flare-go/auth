package role

import (
	"context"
	"go.flare.io/auth/models"
)

type Service interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id uint32) (*models.Role, error)
	Delete(ctx context.Context, roleID uint32) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uint32) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint32) error
	GetRolePermissions(ctx context.Context, roleID uint32) ([]*models.Permission, error)
}
