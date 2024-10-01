package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"goflare.io/auth/driver"
	"goflare.io/auth/models"
	"goflare.io/auth/sqlc"
)

// _ is a type assertion to ensure that the repository implements the Repository interface.
var _ Repository = (*repository)(nil)

// Repository defines the contract for the user repository.
type Repository interface {

	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *models.User) (uint64, error)

	// FindUserByID finds a user by its ID.
	FindUserByID(ctx context.Context, id uint64) (*models.User, error)

	// FindUserByUsername finds a user by its username.
	FindUserByUsername(ctx context.Context, username string) (*models.User, error)

	// FindUserByEmail finds a user by its email.
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)

	// FindUserByFirebaseUID retrieves a user based on their Firebase UID from the database.
	FindUserByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error)

	// AssignRoleToUserWithTx assigns a role to a user.
	AssignRoleToUserWithTx(ctx context.Context, userID, roleID uint64) error

	// FindUserRoles retrieves a user's roles.
	FindUserRoles(ctx context.Context, userID uint64) ([]*models.Role, error)

	// ListAllUsers lists all users.
	ListAllUsers(ctx context.Context) ([]*models.User, error)
}

// repository is the implementation of the Repository interface.
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

// CreateUser creates a new user.
func (r *repository) CreateUser(ctx context.Context, user *models.User) (uint64, error) {
	return sqlc.New(r.conn).CreateUser(ctx, sqlc.CreateUserParams{
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Email:        user.Email,
		FirebaseUid:  &user.FirebaseUID,
		Provider:     sqlc.ProviderType(user.Provider),
	})
}

// FindUserByID finds a user by its ID.
func (r *repository) FindUserByID(ctx context.Context, id uint64) (*models.User, error) {
	if id == 0 {
		return nil, errors.New("id is required")
	}

	sqlcUser, err := sqlc.New(r.conn).FindUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return new(models.User).ConvertFromSQLCUser(sqlcUser), nil
}

// FindUserByUsername finds a user by their username.
// Returns a user model or an error if the user is not found or an error occurs.
func (r *repository) FindUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	sqlcUser, err := sqlc.New(r.conn).FindUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return new(models.User).ConvertFromSQLCUser(sqlcUser), nil
}

// FindUserByEmail finds a user by their email.
// Returns a user model or an error if the user is not found or an error occurs.
func (r *repository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	sqlcUser, err := sqlc.New(r.conn).FindUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return new(models.User).ConvertFromSQLCUser(sqlcUser), nil
}

// FindUserByFirebaseUID retrieves a user based on their Firebase UID from the database.
func (r *repository) FindUserByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error) {
	if firebaseUID == "" {
		return nil, errors.New("firebaseUID is required")
	}

	sqlcUser, err := sqlc.New(r.conn).FindUserByFirebaseUID(ctx, &firebaseUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by firebaseUID: %w", err)
	}

	return new(models.User).ConvertFromSQLCUser(sqlcUser), nil
}

// AssignRoleToUserWithTx assigns a role to a user.
func (r *repository) AssignRoleToUserWithTx(ctx context.Context, userID, roleID uint64) error {
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		}
	}()

	queries := sqlc.New(r.conn).WithTx(tx)
	err = queries.AssignRoleToUser(ctx, sqlc.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindUserRoles retrieves a user's roles.
func (r *repository) FindUserRoles(ctx context.Context, userID uint64) ([]*models.Role, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}

	sqlcRoles, err := sqlc.New(r.conn).GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's roles: %w", err)
	}

	roles := make([]*models.Role, len(sqlcRoles))
	for i, role := range sqlcRoles {
		roles[i] = new(models.Role).ConvertFromSQLCRole(role)
	}

	return roles, nil
}

// ListAllUsers lists all users.
func (r *repository) ListAllUsers(ctx context.Context) ([]*models.User, error) {
	sqlcUsers, err := sqlc.New(r.conn).ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*models.User, len(sqlcUsers))
	for i, user := range sqlcUsers {
		users[i] = new(models.User).ConvertFromSQLCUser(user)
	}

	return users, nil
}
