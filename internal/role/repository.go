package role

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"goflare.io/auth/internal/models"
	"goflare.io/auth/internal/sqlc"
	"goflare.io/nexus/driver"
)

// _ is a type assertion to ensure that the repository implements the Repository interface.
var _ Repository = (*repository)(nil)

// Repository is the interface for the role repository.
type Repository interface {
	CreateRole(ctx context.Context, role *models.Role) error
	FindRoleByID(ctx context.Context, id uint64) (*models.Role, error)
	DeleteRole(ctx context.Context, roleID uint64) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint64) error
	FindRolePermissions(ctx context.Context, roleID uint64) ([]*models.Permission, error)
	ListAllRoles(ctx context.Context) ([]*models.Role, error)
}

// repository is the implementation of the Repository interface.s
type repository struct {
	conn   driver.PostgresPool
	logger *zap.Logger
}

// NewRepository creates a new repository.
func NewRepository(conn driver.PostgresPool, logger *zap.Logger) Repository {
	return &repository{
		conn:   conn,
		logger: logger,
	}
}

// CreateRole creates a new role.
func (r *repository) CreateRole(ctx context.Context, role *models.Role) error {

	return sqlc.New(r.conn).CreateRole(ctx, sqlc.CreateRoleParams{
		Name:        role.Name,
		Description: &role.Description,
	})
}

// FindRoleByID finds a role by ID.
func (r *repository) FindRoleByID(ctx context.Context, id uint64) (*models.Role, error) {

	if id == 0 {
		r.logger.Error("id is required")
		return nil, errors.New("id is required")
	}

	sqlcRole, err := sqlc.New(r.conn).GetRoleByID(ctx, id)
	if err != nil {
		r.logger.Error("error getting role", zap.Error(err))
		return nil, errors.New("error getting role")
	}

	return new(models.Role).ConvertFromSQLCRole(sqlcRole), err
}

// DeleteRole deletes a role by ID.
func (r *repository) DeleteRole(ctx context.Context, roleID uint64) error {

	return sqlc.New(r.conn).DeleteRole(ctx, roleID)
}

// AssignPermissionToRole assigns a permission to a role.
func (r *repository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error {

	return sqlc.New(r.conn).AssignPermissionToRole(ctx, sqlc.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

// RemovePermissionFromRole removes a permission from a role.
func (r *repository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint64) error {

	return sqlc.New(r.conn).RemovePermissionFromRole(ctx, sqlc.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

// FindRolePermissions finds the permissions for a role.
func (r *repository) FindRolePermissions(ctx context.Context, roleID uint64) ([]*models.Permission, error) {

	if roleID == 0 {
		r.logger.Error("id is required")
		return nil, errors.New("id is required")
	}

	sqlcPermissions, err := sqlc.New(r.conn).GetRolePermissions(ctx, roleID)
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

// ListAllRoles lists all roles.
func (r *repository) ListAllRoles(ctx context.Context) ([]*models.Role, error) {

	sqlcRoles, err := sqlc.New(r.conn).ListRoles(ctx)
	if err != nil {
		r.logger.Error("error listing roles", zap.Error(err))
		return nil, errors.New("error listing roles")
	}

	roles := make([]*models.Role, 0, len(sqlcRoles))
	for _, sqlcRole := range sqlcRoles {
		roles = append(roles, new(models.Role).ConvertFromSQLCRole(sqlcRole))
	}

	return roles, nil
}
