package auth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"firebase.google.com/go/v4/auth"
	"goflare.io/auth/firebaseauth"

	"github.com/casbin/casbin/v2"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ed25519"

	"goflare.io/auth/models"
	"goflare.io/auth/models/enum"
	"goflare.io/auth/permission"
	"goflare.io/auth/role"
	"goflare.io/auth/user"
)

type Authentication interface {
	LoadPolicy(ctx context.Context) error
	GeneratePASETOToken(userID uint32) (*models.PASETOToken, error)
	ValidatePASETOToken(tokenString string) (*models.Claims, error)
	RevokePASETOToken(tokenString string) error
	RefreshPASETOToken(refreshToken string) (*models.PASETOToken, error)
	Register(ctx context.Context, username, password, email, phone string) (*models.PASETOToken, error)
	Login(ctx context.Context, email, password string) (*models.PASETOToken, error)
	RegisterWithFirebase(ctx context.Context, username, password, email, phone string) (*models.FirebaseToken, error)
	LoginWithFirebase(ctx context.Context, email, password string) (*models.FirebaseToken, error)
	OAuthLoginWithFirebase(ctx context.Context, provider, idToken string) (*models.FirebaseToken, error)
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
	publicKey      string
	privateKey     string
	user           user.Service
	role           role.Service
	permission     permission.Service
	enforcer       *casbin.Enforcer
	FirebaseClient *firebaseauth.Client
	mu             sync.RWMutex
}

func NewAuthentication(
	user user.Service,
	role role.Service,
	permission permission.Service,
	enforcer *casbin.Enforcer,
	client *firebaseauth.Client,
) Authentication {
	return &AuthenticationImpl{
		user:           user,
		role:           role,
		permission:     permission,
		FirebaseClient: client,
		enforcer:       enforcer,
	}
}

func (s *AuthenticationImpl) LoadPolicy(ctx context.Context) error {
	roleModels, err := s.role.ListAllRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	var errs []error
	for _, roleModel := range roleModels {
		permissions, err := s.role.GetRolePermissions(ctx, roleModel.ID)
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

	// 加載所有用戶和角色
	userModels, err := s.user.ListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	for _, userModel := range userModels {
		userRoles, err := s.user.GetUserRoles(ctx, uint32(userModel.ID))
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

func (s *AuthenticationImpl) GeneratePASETOToken(userID uint32) (*models.PASETOToken, error) {

	expiration := time.Now().Add(120 * time.Minute)

	token := paseto.NewV2()

	payload := models.Claims{
		UserID:     userID,
		Expiration: expiration.Unix(),
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

func (s *AuthenticationImpl) ValidatePASETOToken(tokenString string) (*models.Claims, error) {

	token := paseto.NewV2()

	publicKeyBase64 := s.publicKey
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

func (s *AuthenticationImpl) RevokePASETOToken(tokenString string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *AuthenticationImpl) RefreshPASETOToken(refreshToken string) (*models.PASETOToken, error) {

	token := paseto.NewV2()
	publicKeyBase64 := s.publicKey
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}

	publicKey := ed25519.PublicKey(publicKeyBytes)

	payload := models.Claims{}
	if err = token.Verify(refreshToken, publicKey, &payload, nil); err != nil {
		return nil, err
	}

	return s.GeneratePASETOToken(payload.UserID)
}

func (s *AuthenticationImpl) Register(ctx context.Context, username, password, email, phone string) (*models.PASETOToken, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userModel := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		Phone:        phone,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if _, err = s.user.Create(ctx, userModel); err != nil {
		return nil, err
	}

	return s.GeneratePASETOToken(uint32(userModel.ID))
}

func (s *AuthenticationImpl) Login(ctx context.Context, email, password string) (*models.PASETOToken, error) {

	userModel, err := s.user.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	return s.GeneratePASETOToken(uint32(userModel.ID))
}

func (s *AuthenticationImpl) RegisterWithFirebase(ctx context.Context, username, password, email, phone string) (*models.FirebaseToken, error) {

	// 使用 Firebase Auth 創建用戶
	params := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Password(password).
		DisplayName(username).
		PhoneNumber(phone).
		Disabled(false)

	userRecord, err := s.FirebaseClient.Auth.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 插入用戶到 Postgres
	userModel := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		Phone:        phone,
		FirebaseUID:  userRecord.UID,
		Provider:     "email",
		DisplayName:  userRecord.DisplayName,
		PhotoURL:     userRecord.PhotoURL,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if _, err = s.user.Create(ctx, userModel); err != nil {
		if err = s.FirebaseClient.Auth.DeleteUser(ctx, userRecord.UID); err != nil {
			return nil, err
		}
		return nil, err
	}

	// 生成 Firebase 自定義 Token
	customToken, err := s.FirebaseClient.Auth.CustomToken(ctx, userRecord.UID)
	if err != nil {
		return nil, err
	}

	return &models.FirebaseToken{
		Token: customToken,
	}, nil
}

func (s *AuthenticationImpl) LoginWithFirebase(ctx context.Context, email, password string) (*models.FirebaseToken, error) {

	// 從 Postgres 根據 email 獲取用戶
	userModel, err := s.user.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// 比較密碼
	if err = bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	// 生成 Firebase 自定義 Token
	customToken, err := s.FirebaseClient.Auth.CustomToken(ctx, userModel.FirebaseUID)
	if err != nil {
		return nil, err
	}

	return &models.FirebaseToken{
		Token: customToken,
	}, nil
}

func (s *AuthenticationImpl) OAuthLoginWithFirebase(ctx context.Context, provider, idToken string) (*models.FirebaseToken, error) {

	// 驗證 ID Token 並取得 Firebase 用戶
	decodedToken, err := s.FirebaseClient.Auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid ID token: %w", err)
	}

	userRecord, err := s.FirebaseClient.Auth.GetUser(ctx, decodedToken.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user from Firebase: %w", err)
	}

	// 從 Postgres 根據 Firebase UID 獲取用戶
	userModel, err := s.user.GetByFirebaseUID(ctx, userRecord.UID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 用戶不存在，創建新用戶
			userModel = &models.User{
				Username:     userRecord.DisplayName,
				Email:        userRecord.Email,
				Phone:        userRecord.PhoneNumber,
				FirebaseUID:  userRecord.UID,
				Provider:     strings.ToLower(provider), // 確保 provider 小寫
				DisplayName:  userRecord.DisplayName,
				PhotoURL:     userRecord.PhotoURL,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
				LastSignInAt: time.Now(),
			}

			userID, err := s.user.Create(ctx, userModel)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}

			userModel.ID = int(userID)
		} else {
			return nil, fmt.Errorf("database error: %w", err)
		}
	} else {
		// 用戶已存在，更新最後登入時間
		err = s.user.UpdateLastSignIn(ctx, uint32(userModel.ID))
		if err != nil {
			return nil, fmt.Errorf("failed to update last sign-in time: %w", err)
		}
	}

	// 確保使用者的 provider 一致
	if strings.ToLower(userModel.Provider) != strings.ToLower(provider) {
		return nil, fmt.Errorf("provider mismatch")
	}

	// 生成 Firebase 自定義 Token
	customToken, err := s.FirebaseClient.Auth.CustomToken(ctx, userModel.FirebaseUID)
	if err != nil {
		return nil, fmt.Errorf("failed to create custom token: %w", err)
	}

	return &models.FirebaseToken{
		Token: customToken,
	}, nil
}

func (s *AuthenticationImpl) Logout(token string) error {

	return s.RevokePASETOToken(token)
}

func (s *AuthenticationImpl) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {

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

func (s *AuthenticationImpl) DeleteRole(ctx context.Context, roleID uint32) error {

	// First, remove all references to this role in the Casbin enforcer
	if _, err := s.enforcer.DeleteRole(strconv.Itoa(int(roleID))); err != nil {
		return err
	}

	// Then delete the role from the database
	return s.role.Delete(ctx, roleID)
}

func (s *AuthenticationImpl) AssignRoleToUser(ctx context.Context, userID, roleID uint32) error {

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
	if _, err = s.enforcer.AddGroupingPolicy(userModel.Username, roleModel.Name); err != nil {
		return err
	}

	return nil
}

func (s *AuthenticationImpl) RemoveRoleFromUser(ctx context.Context, userID, roleID uint32) error {

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

func (s *AuthenticationImpl) CreatePermission(ctx context.Context, name, description string, resource enum.ResourceType, action enum.ActionType) (*models.Permission, error) {

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

func (s *AuthenticationImpl) DeletePermission(ctx context.Context, permissionID uint32) error {

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

func (s *AuthenticationImpl) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint32) error {

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

func (s *AuthenticationImpl) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint32) error {

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

func (s *AuthenticationImpl) CheckPermission(ctx context.Context, userID uint32, resource enum.ResourceType, action enum.ActionType) (bool, error) {

	userModel, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	return s.enforcer.Enforce(userModel.Username, string(resource), string(action))
}

func (s *AuthenticationImpl) GetUserRoles(ctx context.Context, userID uint32) ([]string, error) {

	userModel, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.enforcer.GetRolesForUser(userModel.Username)
}

func (s *AuthenticationImpl) GetRolePermissions(ctx context.Context, roleID uint32) ([]*models.Permission, error) {

	permissions, err := s.role.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}
