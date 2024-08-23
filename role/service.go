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
	ListAllRoles(ctx context.Context) ([]*models.Role, error)
}

type service struct {
	repo Repository
}

func NewService(
	repo Repository,
) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, role *models.Role) error {

	return s.repo.Create(ctx, role)
}

func (s *service) GetByID(ctx context.Context, id uint32) (*models.Role, error) {

	return s.repo.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, roleID uint32) error {

	return s.repo.Delete(ctx, roleID)
}

func (s *service) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint32) error {

	return s.repo.AssignPermissionToRole(ctx, roleID, permissionID)
}

func (s *service) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint32) error {

	return s.repo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

func (s *service) GetRolePermissions(ctx context.Context, roleID uint32) ([]*models.Permission, error) {

	return s.repo.GetRolePermissions(ctx, roleID)
}

func (s *service) ListAllRoles(ctx context.Context) ([]*models.Role, error) {

	return s.repo.ListAllRoles(ctx)
}
