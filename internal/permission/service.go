package permission

import (
	"context"

	"goflare.io/auth/internal/models"
)

// Service is the interface for the permission service.
type Service interface {
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uint64) (*models.Permission, error)
	Delete(ctx context.Context, id uint64) error
}

// service is the implementation of the Service interface.
type service struct {
	repo Repository
}

// NewService creates a new service.
func NewService(
	repo Repository,
) Service {
	return &service{
		repo: repo,
	}
}

// Create creates a new permission.
func (s *service) Create(ctx context.Context, permission *models.Permission) error {

	return s.repo.Create(ctx, permission)
}

// GetByID gets a permission by ID.
func (s *service) GetByID(ctx context.Context, id uint64) (*models.Permission, error) {

	return s.repo.GetByID(ctx, id)
}

// Delete deletes a permission by ID.
func (s *service) Delete(ctx context.Context, id uint64) error {

	return s.repo.Delete(ctx, id)
}
