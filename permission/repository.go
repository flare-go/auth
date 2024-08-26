package permission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"goflare.io/auth/driver"
	"goflare.io/auth/models"
	"goflare.io/auth/sqlc"
)

var _ Repository = (*repository)(nil)

type Repository interface {
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uint32) (*models.Permission, error)
	Delete(ctx context.Context, id uint32) error
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

func (r *repository) Create(ctx context.Context, permission *models.Permission) error {

	return r.queries.CreatePermission(ctx, sqlc.CreatePermissionParams{
		Name:        permission.Name,
		Description: &permission.Description,
		Resource:    sqlc.ResourceType(permission.Resource),
		Action:      sqlc.ActionType(permission.Action),
	})
}

func (r *repository) GetByID(ctx context.Context, id uint32) (*models.Permission, error) {

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

func (r *repository) Delete(ctx context.Context, id uint32) error {

	if id == 0 {
		r.logger.Error("id cannot be zero")
		return errors.New("permission id cannot be 0")
	}

	return r.queries.DeletePermission(ctx, id)
}
