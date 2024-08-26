package permission

import (
	"context"
	"goflare.io/auth/models"
)

type Service interface {
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uint32) (*models.Permission, error)
	Delete(ctx context.Context, id uint32) error
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

func (s *service) Create(ctx context.Context, permission *models.Permission) error {

	return s.repo.Create(ctx, permission)
}

func (s *service) GetByID(ctx context.Context, id uint32) (*models.Permission, error) {

	return s.repo.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uint32) error {

	return s.repo.Delete(ctx, id)
}
