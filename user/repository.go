package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go.flare.io/auth/driver"
	"go.flare.io/auth/models"
	"go.flare.io/auth/sqlc"
	"go.uber.org/zap"
)

type Repository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint32) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	AssignRoleToUserWithTx(ctx context.Context, userID, roleID uint32) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error
	GetUserRoles(ctx context.Context, userID uint32) ([]*models.Role, error)
	ListAllUsers(context.Context) ([]*models.User, error)
}

type repository struct {
	db      driver.PostgresPool
	queries sqlc.Querier
	logger  *zap.Logger
}

func NewRepository(conn driver.PostgresPool, logger *zap.Logger) Repository {
	return &repository{
		db:      conn,
		queries: sqlc.New(conn),
		logger:  logger,
	}
}

func (r *repository) Create(ctx context.Context, user *models.User) error {
	return r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Email:        user.Email,
	})
}

func (r *repository) GetByID(ctx context.Context, id uint32) (*models.User, error) {

	if id == 0 {
		r.logger.Error("id is required")
		return nil, errors.New("id is required")
	}

	sqlcUser, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get user by id", zap.Error(err))
		return nil, err
	}

	user := models.NewUser().ConvertFromSQLCUser(sqlcUser)
	user.ID = id

	return user, nil
}

func (r *repository) GetByUsername(ctx context.Context, username string) (*models.User, error) {

	if username == "" {
		r.logger.Error("username is required")
		return nil, errors.New("username is required")
	}

	sqlcUser, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		r.logger.Error("failed to get user by username", zap.Error(err))
		return nil, err
	}

	user := models.NewUser().ConvertFromSQLCUser(sqlcUser)
	user.Username = username

	return user, nil
}

func (r *repository) AssignRoleToUserWithTx(ctx context.Context, userID, roleID uint32) error {

	dbTx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if err = dbTx.Rollback(ctx); err != nil {
				err = fmt.Errorf("rollback failed: %v, original error: %v", err, err)
			}
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			rbErr := dbTx.Rollback(ctx)
			if rbErr != nil {
				err = fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
			}
		} else {
			err = dbTx.Commit(ctx)
		}
	}()

	queries := sqlc.New(r.db).WithTx(dbTx)

	return queries.AssignRoleToUser(ctx, sqlc.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (r *repository) RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error {

	return r.queries.RemoveRoleFromUser(ctx, sqlc.RemoveRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (r *repository) GetUserRoles(ctx context.Context, userID uint32) ([]*models.Role, error) {

	if userID == 0 {
		r.logger.Error("userID is required")
		return nil, errors.New("userID is required")
	}

	sqlcRoles, err := r.queries.GetUserRoles(ctx, userID)
	if err != nil {
		r.logger.Error("failed to get user's roles", zap.Error(err))
		return nil, err
	}

	roles := make([]*models.Role, len(sqlcRoles))
	for _, role := range sqlcRoles {
		roles = append(roles, models.NewRole().ConvertFromSQLCRole(role))
	}

	return roles, nil
}

func (r *repository) ListAllUsers(ctx context.Context) ([]*models.User, error) {

	sqlcUsers, err := r.queries.ListUsers(ctx)
	if err != nil {
		r.logger.Error("failed to list users", zap.Error(err))
		return nil, err
	}

	users := make([]*models.User, len(sqlcUsers))
	for _, user := range sqlcUsers {
		users = append(users, models.NewUser().ConvertFromSQLCUser(user))
	}

	return users, nil
}
