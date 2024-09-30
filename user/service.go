package user

import (
	"context"
	"goflare.io/auth/models"
)

type Service interface {
	Create(ctx context.Context, user *models.User) (uint32, error)
	GetByID(ctx context.Context, id uint32) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error)
	UpdateLastSignIn(ctx context.Context, userID uint32) error
	AssignRoleToUserWithTx(ctx context.Context, userID, roleID uint32) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error
	GetUserRoles(ctx context.Context, userID uint32) ([]*models.Role, error)
	ListAllUsers(context.Context) ([]*models.User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, user *models.User) (uint32, error) {
	return s.repo.Create(ctx, user)
}

func (s *service) GetByID(ctx context.Context, id uint32) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *service) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *service) AssignRoleToUserWithTx(ctx context.Context, userID, roleID uint32) error {
	return s.repo.AssignRoleToUserWithTx(ctx, userID, roleID)
}

func (s *service) RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error {
	return s.repo.RemoveRoleFromUser(ctx, userID, roleID)
}

func (s *service) GetUserRoles(ctx context.Context, userID uint32) ([]*models.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

func (s *service) UpdateLastSignIn(ctx context.Context, userID uint32) error {
	panic("implement me")
}

func (s *service) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error) {
	panic("implement me")
}

func (s *service) ListAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.ListAllUsers(ctx)
}
