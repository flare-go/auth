package firebase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	firebaseauth "firebase.google.com/go/v4/auth"
	"go.uber.org/zap"

	"goflare.io/auth/models"
	"goflare.io/auth/user"
	"golang.org/x/crypto/bcrypt"
)

var _ Service = (*service)(nil)

type Service interface {
	Login(ctx context.Context, email, password string) (*models.FirebaseToken, error)
	Register(ctx context.Context, username, password, email, phone string) (*models.FirebaseToken, error)
	OauthLogin(ctx context.Context, provider, idToken string) (*models.FirebaseToken, error)
}

type service struct {
	userStore      user.Repository
	firebaseclient *firebaseauth.Client
	logger         *zap.Logger
}

func NewService(
	userStore user.Repository,
	firebaseclient *firebaseauth.Client,
	logger *zap.Logger,
) Service {
	return &service{
		userStore:      userStore,
		firebaseclient: firebaseclient,
		logger:         logger,
	}
}

func (s *service) Login(ctx context.Context, email, password string) (*models.FirebaseToken, error) {
	s.logger.Info("Starting login", zap.String("email", email))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	user, err := s.userStore.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get user"))
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("incorrect password")
	}

	customToken, err := s.firebaseclient.CustomToken(ctx, user.FirebaseUID)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to generate Firebase token"))
	}

	return &models.FirebaseToken{Token: customToken}, nil
}

func (s *service) Register(ctx context.Context, username, password, email, phone string) (*models.FirebaseToken, error) {
	s.logger.Info("Starting registration", zap.String("email", email))

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	params := (&firebaseauth.UserToCreate{}).
		Email(email).
		Password(password).
		DisplayName(username).
		PhoneNumber(phone)

	firebaseUser, err := s.firebaseclient.CreateUser(ctx, params)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to create Firebase user"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to hash password"))
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		Phone:        phone,
		FirebaseUID:  firebaseUser.UID,
		Provider:     "email",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if _, err := s.userStore.CreateUser(ctx, user); err != nil {
		// If creating local user fails, delete Firebase user
		if delErr := s.firebaseclient.DeleteUser(ctx, firebaseUser.UID); delErr != nil {
			return nil, errors.Join(delErr, errors.New("failed to delete Firebase user"))
		}
		return nil, errors.Join(err, errors.New("failed to create local user"))
	}

	customToken, err := s.firebaseclient.CustomToken(ctx, firebaseUser.UID)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to generate Firebase token"))
	}

	return &models.FirebaseToken{Token: customToken}, nil
}

func (s *service) OauthLogin(ctx context.Context, provider, idToken string) (*models.FirebaseToken, error) {
	s.logger.Info("Starting OAuth login", zap.String("provider", provider))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Verify ID Token
	decodedToken, err := s.firebaseclient.VerifyIDToken(ctx, idToken)
	if err != nil {
		s.logger.Error("Failed to verify ID Token", zap.Error(err))
		return nil, fmt.Errorf("invalid ID Token: %w", err)
	}

	// Get Firebase user
	firebaseUser, err := s.firebaseclient.GetUser(ctx, decodedToken.UID)
	if err != nil {
		s.logger.Error("Failed to get Firebase user", zap.Error(err))
		return nil, fmt.Errorf("failed to get Firebase user: %w", err)
	}

	// Check if provider is supported
	if !isSupportedProvider(provider) {
		s.logger.Warn("Unsupported OAuth provider", zap.String("provider", provider))
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Get user from database
	user, err := s.userStore.FindUserByFirebaseUID(ctx, firebaseUser.UID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new user
			user = &models.User{
				Username:    firebaseUser.DisplayName,
				Email:       firebaseUser.Email,
				FirebaseUID: firebaseUser.UID,
				Provider:    strings.ToLower(provider),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if _, err := s.userStore.CreateUser(ctx, user); err != nil {
				s.logger.Error("Failed to create user", zap.Error(err))
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			s.logger.Info("Created new user", zap.String("email", user.Email))
		} else {
			s.logger.Error("Failed to query user", zap.Error(err))
			return nil, fmt.Errorf("failed to query user: %w", err)
		}
	} else {
		// // Update user's last login time
		// if err := a.userStore.UpdateLastLogin(ctx, user.ID); err != nil {
		// 	a.logger.Warn("Failed to update last login time", zap.Error(err))
		// }
	}

	// Check if provider matches
	if strings.ToLower(user.Provider) != strings.ToLower(provider) {
		s.logger.Warn("Provider mismatch", zap.String("expected", user.Provider), zap.String("actual", provider))
		return nil, fmt.Errorf("provider mismatch")
	}

	// Generate custom token
	customToken, err := s.firebaseclient.CustomToken(ctx, user.FirebaseUID)
	if err != nil {
		s.logger.Error("Failed to generate Firebase token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate Firebase token: %w", err)
	}

	s.logger.Info("OAuth login successful", zap.String("email", user.Email))
	return &models.FirebaseToken{Token: customToken}, nil
}

// isSupportedProvider checks if the provider is supported.
func isSupportedProvider(provider string) bool {
	supportedProviders := []string{"google", "facebook", "twitter", "github"}
	provider = strings.ToLower(provider)
	for _, p := range supportedProviders {
		if provider == p {
			return true
		}
	}
	return false
}
