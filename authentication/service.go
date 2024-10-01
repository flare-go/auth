package authentication

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"

	"goflare.io/auth/config"
	"goflare.io/auth/models"
	"goflare.io/auth/models/enum"
	"goflare.io/auth/token"
	"goflare.io/auth/user"
)

// _ is used to ensure that *service implements the Service interface at compile time.
var _ Service = (*service)(nil)

// Service represents the set of methods for user authentication and authorization.
type Service interface {
	// Login logs in a user with email and password.
	Login(ctx context.Context, email, password string) (*models.PASETOToken, error)
	// Logout logs out a user with a token.
	Logout(ctx context.Context, token string) error
	// Register registers a new user with username, password, email, and phone.
	Register(ctx context.Context, username, password, email, phone string) (*models.PASETOToken, error)
	// ValidateToken validates a token.
	ValidateToken(token string) (uint64, error)
	// CheckPermission checks if a user has a permission for a resource and action.
	CheckPermission(ctx context.Context, userID uint64, resource enum.ResourceType, action enum.ActionType) (bool, error)
}

// service represents the core service implementation for user management and authorization.
// It integrates various components such as the user repository, token manager, policy enforcer, and logger.
type service struct {
	userStore    user.Repository
	tokenManager token.Manager
	enforcer     *casbin.Enforcer
	logger       *zap.Logger
}

// NewService creates a new instance of Service with a provided user repository, configuration, enforcer, and logger.
func NewService(userStore user.Repository, config *config.Config, enforcer *casbin.Enforcer, logger *zap.Logger) Service {

	tokenManager := token.NewPasetoManager(config.Paseto.PublicSecretKey, config.Paseto.PrivateSecretKey, config.Paseto.TokenExpirationTime)
	return &service{
		userStore:    userStore,
		tokenManager: tokenManager,
		enforcer:     enforcer,
		logger:       logger,
	}
}

// Login authenticates a user with the provided email and password,
// returning a PASETO token upon successful authentication.
func (s *service) Login(ctx context.Context, email, password string) (*models.PASETOToken, error) {
	s.logger.Info("login user", zap.String("email", email))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := s.userStore.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user does not exist")
		}
		return nil, errors.Join(err, errors.New("failed to get user"))
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("incorrect password")
	}

	return s.tokenManager.GenerateToken(user.ID)
}

// Logout revokes the provided token, effectively logging the user out.
func (s *service) Logout(ctx context.Context, token string) error {
	s.logger.Info("logout user", zap.String("token", token))
	return s.tokenManager.RevokeToken(token)
}

// Register registers a new user, hashes the password, stores the user, and generates a PASETO token.
func (s *service) Register(ctx context.Context, username, password, email, phone string) (*models.PASETOToken, error) {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to hash password"))
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		Phone:        phone,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if _, err := s.userStore.CreateUser(ctx, user); err != nil {
		return nil, errors.Join(err, errors.New("failed to create user"))
	}

	return s.tokenManager.GenerateToken(user.ID)
}

// ValidateToken validates the given token and returns the associated user ID if valid, otherwise an error.
func (s *service) ValidateToken(token string) (uint64, error) {
	s.logger.Info("validate token", zap.String("token", token))
	return s.tokenManager.ValidateToken(token)
}

// CheckPermission verifies if the user has permission to perform a specific action on a resource.
func (s *service) CheckPermission(ctx context.Context, userID uint64, resource enum.ResourceType, action enum.ActionType) (bool, error) {
	s.logger.Info("check permission", zap.Uint64("userID", userID), zap.Any("resource", resource), zap.Any("action", action))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := s.userStore.FindUserByID(ctx, userID)
	if err != nil {
		return false, errors.Join(err, errors.New("failed to get user"))
	}

	return s.enforcer.Enforce(user.Username, string(resource), string(action))
}
