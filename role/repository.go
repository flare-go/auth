package role

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"goflare.io/auth/driver"
	"goflare.io/auth/models"
	"goflare.io/auth/sqlc"
)

type Repository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id uint32) (*models.Role, error)
	Delete(ctx context.Context, roleID uint32) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uint32) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint32) error
	GetRolePermissions(ctx context.Context, roleID uint32) ([]*models.Permission, error)
	ListAllRoles(ctx context.Context) ([]*models.Role, error)
}

type repository struct {
	queries sqlc.Querier
	logger  *zap.Logger
}

func NewRepository(conn driver.PostgresPool, logger *zap.Logger) Repository {
	return &repository{
		queries: sqlc.New(conn),
		logger:  logger,
	}
}

func (r *repository) Create(ctx context.Context, role *models.Role) error {

	return r.queries.CreateRole(ctx, sqlc.CreateRoleParams{
		Name:        role.Name,
		Description: &role.Description,
	})
}

func (r *repository) GetByID(ctx context.Context, id uint32) (*models.Role, error) {

	if id == 0 {
		r.logger.Error("id is required")
		return nil, errors.New("id is required")
	}

	sqlcRole, err := r.queries.GetRoleByID(ctx, id)
	if err != nil {
		r.logger.Error("error getting role", zap.Error(err))
		return nil, errors.New("error getting role")
	}

	return models.NewRole().ConvertFromSQLCRole(sqlcRole), err
}

func (r *repository) Delete(ctx context.Context, roleID uint32) error {

	return r.queries.DeleteRole(ctx, roleID)
}

func (r *repository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint32) error {

	return r.queries.AssignPermissionToRole(ctx, sqlc.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (r *repository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint32) error {

	return r.queries.RemovePermissionFromRole(ctx, sqlc.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (r *repository) GetRolePermissions(ctx context.Context, roleID uint32) ([]*models.Permission, error) {

	if roleID == 0 {
		r.logger.Error("id is required")
		return nil, errors.New("id is required")
	}

	sqlcPermissions, err := r.queries.GetRolePermissions(ctx, roleID)
	if err != nil {
		r.logger.Error("error getting role permissions", zap.Error(err))
		return nil, errors.New("error getting role permissions")
	}

	permissions := make([]*models.Permission, 0, len(sqlcPermissions))

	for _, sqlcPermission := range sqlcPermissions {
		permissions = append(permissions, models.NewPermission().ConvertFromSQLCPermission(sqlcPermission))
	}

	return permissions, nil
}

func (r *repository) ListAllRoles(ctx context.Context) ([]*models.Role, error) {

	sqlcRoles, err := r.queries.ListRoles(ctx)
	if err != nil {
		r.logger.Error("error listing roles", zap.Error(err))
		return nil, errors.New("error listing roles")
	}

	roles := make([]*models.Role, 0, len(sqlcRoles))
	for _, sqlcRole := range sqlcRoles {
		roles = append(roles, models.NewRole().ConvertFromSQLCRole(sqlcRole))
	}

	return roles, nil
}
