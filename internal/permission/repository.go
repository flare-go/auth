package permission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"goflare.io/auth/internal/driver"
	"goflare.io/auth/internal/models"
	"goflare.io/auth/internal/sqlc"
)

// _ is a type assertion to ensure that the repository implements the Repository interface.
var _ Repository = (*repository)(nil)

// Repository is the interface for the permission repository.
type Repository interface {
	// Create creates a new permission.
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uint64) (*models.Permission, error)
	Delete(ctx context.Context, id uint64) error
}

// repository is the implementation of the Repository interface.
type repository struct {
	queries sqlc.Querier
	logger  *zap.Logger
}

// NewRepository creates a new repository.
func NewRepository(conn driver.PostgresPool, logger *zap.Logger) Repository {

	return &repository{
		queries: sqlc.New(conn),
		logger:  logger,
	}
}

// Create creates a new permission.
func (r *repository) Create(ctx context.Context, permission *models.Permission) error {

	return r.queries.CreatePermission(ctx, sqlc.CreatePermissionParams{
		Name:        permission.Name,
		Description: &permission.Description,
		Resource:    sqlc.ResourceType(permission.Resource),
		Action:      sqlc.ActionType(permission.Action),
	})
}

// GetByID gets a permission by ID.
func (r *repository) GetByID(ctx context.Context, id uint64) (*models.Permission, error) {

	if id == 0 {
		r.logger.Error("id cannot be zero")
		return nil, errors.New("permission id cannot be 0")
	}

	sqlcPermission, err := r.queries.GetPermissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission from database: %w", err)
	}

	return models.NewPermission().ConvertFromSQLCPermission(sqlcPermission), nil
}

// Delete deletes a permission by ID.
func (r *repository) Delete(ctx context.Context, id uint64) error {

	if id == 0 {
		r.logger.Error("id cannot be zero")
		return errors.New("permission id cannot be 0")
	}

	return r.queries.DeletePermission(ctx, id)
}
