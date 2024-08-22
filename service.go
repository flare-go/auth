package auth

import (
	"context"
	"encoding/base64"
	"github.com/casbin/casbin/v3"
	"github.com/o1egl/paseto"
	"go.flare.io/auth/driver"
	"go.flare.io/auth/models"
	"go.flare.io/auth/models/enum"
	"go.flare.io/auth/permission"
	"go.flare.io/auth/role"
	"go.flare.io/auth/user"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ed25519"
	"strconv"
	"sync"
	"time"
)

type Service interface {
}

type service struct {
	publicKey  string
	privateKey string
	db         driver.PostgresPool
	user       user.Service
	role       role.Service
	permission permission.Service
	enforcer   *casbin.Enforcer
	mu         sync.RWMutex
}

func (s *service) GeneratePASETOToken(userID uint32) (*models.PASETOToken, error) {

	expiration := time.Now().Add(120 * time.Minute)

	token := paseto.NewV2()

	payload := map[string]any{
		"UserID":     userID,
		"Expiration": expiration.Unix(),
	}
	privateKeyBase64 := s.privateKey
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, err
	}

	privateKey := ed25519.PrivateKey(privateKeyBytes)
	signed, err := token.Sign(privateKey, payload, nil)
	if err != nil {
		return nil, err
	}

	return &models.PASETOToken{
		Token:     signed,
		ExpiresAt: expiration,
	}, nil
}

func (s *service) ValidatePASETOToken(tokenString string) (*models.Claims, error) {

	token := paseto.NewV2()

	publicKeyBase64 := s.publicKey
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}

	publicKey := ed25519.PublicKey(publicKeyBytes)

	var payload map[string]any
	if err = token.Verify(tokenString, publicKey, &payload, nil); err != nil {
		return nil, err
	}

	expiration := time.Unix(int64(payload["Expiration"].(float64)), 0)
	if time.Now().After(expiration) {
		return nil, err
	}

	userID := payload["UserID"].(uint32)

	return &models.Claims{
		UserID:    userID,
		ExpiresAt: expiration,
	}, nil
}

func (s *service) RevokePASETOToken(tokenString string) error {

	token := paseto.NewV2()
	publicKeyBase64 := s.publicKey
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return err
	}

	publicKey := ed25519.PublicKey(publicKeyBytes)

	var payload map[string]any
	if err = token.Verify(tokenString, publicKey, &payload, nil); err != nil {
		return err
	}

	//TODO:
	//在這裡可以將令牌添加到一個"黑名單"或"撤銷列表"中,
	//這樣即使令牌還沒有過期,它也不能再被使用。
	//具體的實現取決於您的系統設計。
	//一個簡單的方法是在數據庫中創建一個"revoked_tokens"表,
	//然後在這裡將令牌插入到該表中。
	//在驗證令牌時,您需要檢查令牌是否在這個表中。

	return nil
}

func (s *service) RefreshPASETOToken(refreshToken string) (*models.PASETOToken, error) {

	token := paseto.NewV2()
	publicKeyBase64 := s.publicKey
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}

	publicKey := ed25519.PublicKey(publicKeyBytes)

	var payload map[string]any
	if err = token.Verify(refreshToken, publicKey, &payload, nil); err != nil {
		return nil, err
	}

	userID := payload["UserID"].(uint32)

	return s.GeneratePASETOToken(userID)
}

func (s *service) Register(ctx context.Context, username, password, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.user.Create(ctx, &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
}

func (s *service) Login(ctx context.Context, username, password string) (*models.PASETOToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userModel, err := s.user.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	return s.GeneratePASETOToken(userModel.ID)
}

func (s *service) Logout(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.RevokePASETOToken(token)
}

func (s *service) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	roleModel := &models.Role{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.role.Create(ctx, roleModel); err != nil {
		return nil, err
	}

	return roleModel, nil
}

func (s *service) DeleteRole(ctx context.Context, roleID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// First, remove all references to this role in the Casbin enforcer
	if _, err := s.enforcer.DeleteRole(strconv.Itoa(int(roleID))); err != nil {
		return err
	}

	// Then delete the role from the database
	return s.role.Delete(ctx, roleID)
}

func (s *service) AssignRoleToUser(ctx context.Context, userID, roleID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userModel, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	roleModel, err := s.role.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	if err = s.user.AssignRoleToUserWithTx(ctx, userID, roleID); err != nil {
		return err
	}

	// Add the role to the user in Casbin
	if _, err = s.enforcer.AddRoleForUser(userModel.Username, roleModel.Name); err != nil {
		return err
	}

	return nil
}

func (s *service) RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userModel, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	roleModel, err := s.role.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	if err = s.user.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return err
	}

	// Remove the role from the user in Casbin
	if _, err = s.enforcer.DeleteRoleForUser(userModel.Username, roleModel.Name); err != nil {
		return err
	}

	return nil
}

func (s *service) CreatePermission(ctx context.Context, name, description string, resource enum.ResourceType, action enum.ActionType) (*models.Permission, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	perm := &models.Permission{
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.permission.Create(ctx, perm); err != nil {
		return nil, err
	}

	return perm, nil
}

func (s *service) DeletePermission(ctx context.Context, permissionID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	perm, err := s.permission.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}

	// Remove all policies related to this permission from Casbin
	if _, err = s.enforcer.RemoveFilteredPolicy(1, string(perm.Resource), string(perm.Action)); err != nil {
		return err
	}

	// Delete the permission from the database
	return s.permission.Delete(ctx, permissionID)
}

func (s *service) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	roleModel, err := s.role.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	perm, err := s.permission.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}

	if err = s.role.AssignPermissionToRole(ctx, roleID, permissionID); err != nil {
		return err
	}

	// Add the permission to the role in Casbin
	if _, err = s.enforcer.AddPolicy(roleModel.Name, string(perm.Resource), string(perm.Action)); err != nil {
		return err
	}

	return nil
}

func (s *service) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	roleModel, err := s.role.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	permModel, err := s.permission.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}

	if err = s.role.RemovePermissionFromRole(ctx, roleID, permissionID); err != nil {
		return err
	}

	// Remove the permission from the role in Casbin
	if _, err = s.enforcer.RemovePolicy(roleModel.Name, string(permModel.Resource), string(permModel.Action)); err != nil {
		return err
	}

	return nil
}

func (s *service) CheckPermission(ctx context.Context, userID uint32, resource enum.ResourceType, action enum.ActionType) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userModel, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	return s.enforcer.Enforce(userModel.Username, string(resource), string(action))
}

func (s *service) GetUserRoles(ctx context.Context, userID uint32) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userModel, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.enforcer.GetRolesForUser(userModel.Username)
}

func (s *service) GetRolePermissions(ctx context.Context, roleID uint32) ([]*models.Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permissions, err := s.role.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}
