package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/o1egl/paseto"
	"goflare.io/auth/models"
	"goflare.io/auth/models/enum"
	"goflare.io/auth/permission"
	"goflare.io/auth/role"
	"goflare.io/auth/user"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ed25519"
	"strconv"
	"sync"
	"time"
)

type Authentication interface {
	LoadPolicy() error
	GeneratePASETOToken(userID uint32) (*models.PASETOToken, error)
	ValidatePASETOToken(tokenString string) (*models.Claims, error)
	RevokePASETOToken(tokenString string) error
	RefreshPASETOToken(refreshToken string) (*models.PASETOToken, error)
	Register(ctx context.Context, username, password, email string) (*models.PASETOToken, error)
	Login(ctx context.Context, email, password string) (*models.PASETOToken, error)
	Logout(token string) error
	CreateRole(ctx context.Context, name, description string) (*models.Role, error)
	DeleteRole(ctx context.Context, roleID uint32) error
	AssignRoleToUser(ctx context.Context, userID, roleID uint32) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error
	CreatePermission(ctx context.Context, name, description string, resource enum.ResourceType, action enum.ActionType) (*models.Permission, error)
	DeletePermission(ctx context.Context, permissionID uint32) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uint32) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint32) error
	CheckPermission(ctx context.Context, userID uint32, resource enum.ResourceType, action enum.ActionType) (bool, error)
	GetUserRoles(ctx context.Context, userID uint32) ([]string, error)
	GetRolePermissions(ctx context.Context, roleID uint32) ([]*models.Permission, error)
}

type AuthenticationImpl struct {
	publicKey  string
	privateKey string
	user       user.Service
	role       role.Service
	permission permission.Service
	enforcer   *casbin.Enforcer
	mu         sync.RWMutex
}

func NewAuthentication(
	user user.Service,
	role role.Service,
	permission permission.Service,
	secret models.PasetoSecret,
	enforcer *casbin.Enforcer,
) Authentication {
	return &AuthenticationImpl{
		publicKey:  secret.PasetoPublicKey,
		privateKey: secret.PasetoPrivateKey,
		user:       user,
		role:       role,
		permission: permission,
		enforcer:   enforcer,
	}
}

func (auth *AuthenticationImpl) LoadPolicy() error {
	ctx := context.Background()

	// 加載所有角色和權限
	roleModels, err := auth.role.ListAllRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	for _, roleModel := range roleModels {
		permissions, err := auth.role.GetRolePermissions(ctx, roleModel.ID)
		if err != nil {
			return fmt.Errorf("failed to get permissions for role %s: %w", roleModel.Name, err)
		}

		for _, perm := range permissions {
			if _, err = auth.enforcer.AddPolicy(roleModel.Name, string(perm.Resource), string(perm.Action)); err != nil {
				return fmt.Errorf("failed to add policy for role %s: %w", roleModel.Name, err)
			}
		}
	}

	// 加載所有用戶和角色
	userModels, err := auth.user.ListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	for _, userModel := range userModels {
		userRoles, err := auth.user.GetUserRoles(ctx, userModel.ID)
		if err != nil {
			return fmt.Errorf("failed to get roles for user %s: %w", userModel.Username, err)
		}

		for _, userRole := range userRoles {
			if _, err = auth.enforcer.AddGroupingPolicy(userModel.Username, userRole.Name); err != nil {
				return fmt.Errorf("failed to add grouping policy for user %s: %w", userModel.Username, err)
			}
		}
	}

	return nil
}

func (auth *AuthenticationImpl) GeneratePASETOToken(userID uint32) (*models.PASETOToken, error) {

	expiration := time.Now().Add(120 * time.Minute)

	token := paseto.NewV2()

	//payload := map[string]any{
	//	"UserID":     userID,
	//	"Expiration": expiration.Unix(),
	//}

	payload := models.Claims{
		UserID:     userID,
		Expiration: expiration.Unix(),
	}
	privateKeyBase64 := auth.privateKey
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

func (auth *AuthenticationImpl) ValidatePASETOToken(tokenString string) (*models.Claims, error) {

	token := paseto.NewV2()

	publicKeyBase64 := auth.publicKey
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}

	publicKey := ed25519.PublicKey(publicKeyBytes)

	payload := models.Claims{}
	if err = token.Verify(tokenString, publicKey, &payload, nil); err != nil {
		return nil, err
	}

	//expiration := time.Unix(int64(payload["Expiration"].(float64)), 0)
	if time.Now().After(time.Unix(payload.Expiration, 0)) {
		return nil, err
	}

	return &payload, nil
}

func (auth *AuthenticationImpl) RevokePASETOToken(tokenString string) error {

	token := paseto.NewV2()
	publicKeyBase64 := auth.publicKey
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

func (auth *AuthenticationImpl) RefreshPASETOToken(refreshToken string) (*models.PASETOToken, error) {

	token := paseto.NewV2()
	publicKeyBase64 := auth.publicKey
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}

	publicKey := ed25519.PublicKey(publicKeyBytes)

	payload := models.Claims{}
	if err = token.Verify(refreshToken, publicKey, &payload, nil); err != nil {
		return nil, err
	}

	return auth.GeneratePASETOToken(payload.UserID)
}

func (auth *AuthenticationImpl) Register(ctx context.Context, username, password, email string) (*models.PASETOToken, error) {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id, err := auth.user.Create(ctx, &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})

	if err != nil {
		return nil, err
	}

	return auth.GeneratePASETOToken(id)
}

func (auth *AuthenticationImpl) Login(ctx context.Context, email, password string) (*models.PASETOToken, error) {
	auth.mu.RLock()
	defer auth.mu.RUnlock()

	userModel, err := auth.user.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	return auth.GeneratePASETOToken(userModel.ID)
}

func (auth *AuthenticationImpl) Logout(token string) error {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	return auth.RevokePASETOToken(token)
}

func (auth *AuthenticationImpl) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	roleModel := &models.Role{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := auth.role.Create(ctx, roleModel); err != nil {
		return nil, err
	}

	return roleModel, nil
}

func (auth *AuthenticationImpl) DeleteRole(ctx context.Context, roleID uint32) error {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	// First, remove all references to this role in the Casbin enforcer
	if _, err := auth.enforcer.DeleteRole(strconv.Itoa(int(roleID))); err != nil {
		return err
	}

	// Then delete the role from the database
	return auth.role.Delete(ctx, roleID)
}

func (auth *AuthenticationImpl) AssignRoleToUser(ctx context.Context, userID, roleID uint32) error {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	userModel, err := auth.user.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	roleModel, err := auth.role.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	if err = auth.user.AssignRoleToUserWithTx(ctx, userID, roleID); err != nil {
		return err
	}

	// Add the role to the user in Casbin
	if _, err = auth.enforcer.AddGroupingPolicy(userModel.Username, roleModel.Name); err != nil {
		return err
	}

	return nil
}

func (auth *AuthenticationImpl) RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	userModel, err := auth.user.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	roleModel, err := auth.role.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	if err = auth.user.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return err
	}

	// Remove the role from the user in Casbin
	if _, err = auth.enforcer.DeleteRoleForUser(userModel.Username, roleModel.Name); err != nil {
		return err
	}

	return nil
}

func (auth *AuthenticationImpl) CreatePermission(ctx context.Context, name, description string, resource enum.ResourceType, action enum.ActionType) (*models.Permission, error) {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	perm := &models.Permission{
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := auth.permission.Create(ctx, perm); err != nil {
		return nil, err
	}

	return perm, nil
}

func (auth *AuthenticationImpl) DeletePermission(ctx context.Context, permissionID uint32) error {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	perm, err := auth.permission.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}

	// Remove all policies related to this permission from Casbin
	if _, err = auth.enforcer.RemoveFilteredPolicy(1, string(perm.Resource), string(perm.Action)); err != nil {
		return err
	}

	// Delete the permission from the database
	return auth.permission.Delete(ctx, permissionID)
}

func (auth *AuthenticationImpl) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint32) error {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	roleModel, err := auth.role.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	perm, err := auth.permission.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}

	if err = auth.role.AssignPermissionToRole(ctx, roleID, permissionID); err != nil {
		return err
	}

	// Add the permission to the role in Casbin
	if _, err = auth.enforcer.AddPolicy(roleModel.Name, string(perm.Resource), string(perm.Action)); err != nil {
		return err
	}

	return nil
}

func (auth *AuthenticationImpl) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint32) error {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	roleModel, err := auth.role.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	permModel, err := auth.permission.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}

	if err = auth.role.RemovePermissionFromRole(ctx, roleID, permissionID); err != nil {
		return err
	}

	// Remove the permission from the role in Casbin
	if _, err = auth.enforcer.RemovePolicy(roleModel.Name, string(permModel.Resource), string(permModel.Action)); err != nil {
		return err
	}

	return nil
}

func (auth *AuthenticationImpl) CheckPermission(ctx context.Context, userID uint32, resource enum.ResourceType, action enum.ActionType) (bool, error) {
	auth.mu.RLock()
	defer auth.mu.RUnlock()

	userModel, err := auth.user.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	return auth.enforcer.Enforce(userModel.Username, string(resource), string(action))
}

func (auth *AuthenticationImpl) GetUserRoles(ctx context.Context, userID uint32) ([]string, error) {
	auth.mu.RLock()
	defer auth.mu.RUnlock()

	userModel, err := auth.user.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return auth.enforcer.GetRolesForUser(userModel.Username)
}

func (auth *AuthenticationImpl) GetRolePermissions(ctx context.Context, roleID uint32) ([]*models.Permission, error) {
	auth.mu.RLock()
	defer auth.mu.RUnlock()

	permissions, err := auth.role.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}
