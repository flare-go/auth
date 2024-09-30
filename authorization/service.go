package authorization

import (
	"context"
	"errors"
	"fmt"

	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
	"goflare.io/auth/role"
	"goflare.io/auth/user"
)

var _ Service = (*service)(nil)

// Service is the interface for the authorization service.
type Service interface {
	// LoadPolicies loads all policies.
	LoadPolicies(ctx context.Context) error
}

type service struct {
	userStore user.Repository
	roleStore role.Repository
	enforcer  *casbin.Enforcer
	logger    *zap.Logger
}

func NewService(
	userStore user.Repository,
	roleStore role.Repository,
	enforcer *casbin.Enforcer,
	logger *zap.Logger,
) Service {
	return &service{
		userStore: userStore,
		roleStore: roleStore,
		enforcer:  enforcer,
		logger:    logger,
	}
}

// LoadPolicies loads all policies.
func (s *service) LoadPolicies(ctx context.Context) error {
	roleModels, err := s.roleStore.ListAllRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	var errs []error
	for _, roleModel := range roleModels {
		permissions, err := s.roleStore.FindRolePermissions(ctx, roleModel.ID)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get permissions for role %s: %w", roleModel.Name, err))
			continue
		}

		for _, perm := range permissions {
			if _, err = s.enforcer.AddPolicy(roleModel.Name, string(perm.Resource), string(perm.Action)); err != nil {
				errs = append(errs, fmt.Errorf("failed to add policy for role %s: %w", roleModel.Name, err))
			}
		}
	}

	// Load all users and roles
	userModels, err := s.userStore.ListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	for _, userModel := range userModels {
		userRoles, err := s.userStore.FindUserRoles(ctx, uint32(userModel.ID))
		if err != nil {
			return fmt.Errorf("failed to get roles for user %s: %w", userModel.Username, err)
		}

		for _, userRole := range userRoles {
			if _, err = s.enforcer.AddGroupingPolicy(userModel.Username, userRole.Name); err != nil {
				return fmt.Errorf("failed to add grouping policy for user %s: %w", userModel.Username, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
