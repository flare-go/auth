package firebase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"go.uber.org/zap"

	"goflare.io/auth/internal/config"
	"goflare.io/auth/internal/models"
	"goflare.io/auth/internal/user"
)

// Service is the interface for the Firebase service.
var _ Service = (*service)(nil)

// Service is the interface for the Firebase service.
type Service interface {
	// Login logs in a user with email and password.
	Login(ctx context.Context, email, password string) (*models.FirebaseToken, error)
	// Register registers a new user with username, password, email, and phone.
	Register(ctx context.Context, username, password, email, phone string) (*models.FirebaseToken, error)
	// OauthLogin logs in a user with an OAuth provider and an ID token.
	OauthLogin(ctx context.Context, provider, idToken string) (*models.FirebaseToken, error)
	// GetOAuthURL returns the URL for the OAuth provider's login page.
	GetOAuthURL(provider string) (string, error)
	// ExchangeOAuthToken exchanges an OAuth code for a Firebase token.
	ExchangeOAuthToken(ctx context.Context, provider, code string) (*models.FirebaseToken, error)
}

type service struct {
	userStore   user.Repository
	client      *firebaseauth.Client
	logger      *zap.Logger
	oauthConfig *oauth2.Config
	projectID   string
}

// NewService creates a new Firebase service.
func NewService(
	userStore user.Repository,
	config *config.Config,
	logger *zap.Logger,
) (Service, error) {
	opt := option.WithCredentialsFile(config.Firebase.ServiceAccountFilePath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %v", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.Firebase.ClientID,
		ClientSecret: config.Firebase.ClientSecret,
		RedirectURL:  config.Firebase.RedirectURL,
	}

	return &service{
		userStore:   userStore,
		client:      client,
		logger:      logger,
		oauthConfig: oauthConfig,
		projectID:   config.Firebase.ProjectID,
	}, nil
}

// Login logs in a user with email and password.
func (s *service) Login(ctx context.Context, email, password string) (*models.FirebaseToken, error) {
	s.logger.Info("Starting login", zap.String("email", email))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	userModel, err := s.userStore.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get user"))
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("incorrect password")
	}

	customToken, err := s.client.CustomToken(ctx, userModel.FirebaseUID)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to generate Firebase token"))
	}

	return &models.FirebaseToken{Token: customToken}, nil
}

// Register registers a new user with username, password, email, and phone.
func (s *service) Register(ctx context.Context, username, password, email, phone string) (*models.FirebaseToken, error) {
	s.logger.Info("Starting registration", zap.String("email", email))

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	params := (&firebaseauth.UserToCreate{}).
		Email(email).
		Password(password).
		DisplayName(username).
		PhoneNumber(phone)

	firebaseUser, err := s.client.CreateUser(ctx, params)
	if err != nil {
		s.logger.Error("failed create user")
		return nil, errors.Join(err, errors.New("failed to create Firebase user"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to hash password"))
	}

	s.logger.Info("firebaseUser", zap.Any("firebaseUser", firebaseUser))

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		Phone:        phone,
		FirebaseUID:  firebaseUser.UID,
		Provider:     firebaseUser.ProviderID,
	}

	if _, err := s.userStore.CreateUser(ctx, user); err != nil {
		// If creating a local user fails, delete Firebase user
		if delErr := s.client.DeleteUser(ctx, firebaseUser.UID); delErr != nil {
			return nil, errors.Join(delErr, errors.New("failed to delete Firebase user"))
		}
		return nil, errors.Join(err, errors.New("failed to create local user"))
	}

	customToken, err := s.client.CustomToken(ctx, firebaseUser.UID)
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
	decodedToken, err := s.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		s.logger.Error("Failed to verify ID Token", zap.Error(err))
		return nil, fmt.Errorf("invalid ID Token: %w", err)
	}

	// Get Firebase user
	firebaseUser, err := s.client.GetUser(ctx, decodedToken.UID)
	if err != nil {
		s.logger.Error("Failed to get Firebase user", zap.Error(err))
		return nil, fmt.Errorf("failed to get Firebase user: %w", err)
	}

	// Check if the provider is supported
	if !isSupportedProvider(provider) {
		s.logger.Warn("Unsupported OAuth provider", zap.String("provider", provider))
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Get user from a database
	user, err := s.userStore.FindUserByFirebaseUID(ctx, firebaseUser.UID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create a new user
			user = &models.User{
				Username:    firebaseUser.DisplayName,
				Email:       firebaseUser.Email,
				FirebaseUID: firebaseUser.UID,
				Provider:    strings.ToLower(provider),
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
	}
	// } else {
	// 	// // Update user's last login time
	// 	// if err := a.userStore.UpdateLastLogin(ctx, user.ID); err != nil {
	// 	// 	a.logger.Warn("Failed to update last login time", zap.Error(err))
	// 	// }
	// }

	// Check if provider matches
	if !strings.EqualFold(user.Provider, provider) {
		s.logger.Warn("Provider mismatch", zap.String("expected", user.Provider), zap.String("actual", provider))
		return nil, fmt.Errorf("provider mismatch")
	}

	// Generate custom token
	customToken, err := s.client.CustomToken(ctx, user.FirebaseUID)
	if err != nil {
		s.logger.Error("Failed to generate Firebase token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate Firebase token: %w", err)
	}

	s.logger.Info("OAuth login successful", zap.String("email", user.Email))
	return &models.FirebaseToken{Token: customToken}, nil
}

// there isSupportedProvider checks if the provider is supported.
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

// GetOAuthURL returns the URL for the OAuth provider's login page.
func (s *service) GetOAuthURL(provider string) (string, error) {
	if !isSupportedProvider(provider) {
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	state := generateRandomState()

	var scopes []string
	switch provider {
	case "google":
		scopes = []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"}
	case "facebook":
		scopes = []string{"email", "public_profile"}
	case "github":
		scopes = []string{"user:email"}
	case "twitter":
		scopes = []string{"email"}
	}

	s.oauthConfig.Scopes = scopes
	s.oauthConfig.Endpoint = providerEndpoint(provider)

	authURL := s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return authURL, nil
}

// ExchangeOAuthToken exchanges an OAuth code for a Firebase token.
func (s *service) ExchangeOAuthToken(ctx context.Context, provider, code string) (*models.FirebaseToken, error) {
	if !isSupportedProvider(provider) {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange OAuth code: %w", err)
	}

	userInfo, err := s.getUserInfo(ctx, provider, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	firebaseToken, err := s.createOrUpdateFirebaseUser(ctx, userInfo, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create or update Firebase user: %w", err)
	}

	return firebaseToken, nil
}

// getUserInfo gets the user info from the OAuth provider.
func (s *service) getUserInfo(ctx context.Context, provider, accessToken string) (map[string]interface{}, error) {
	var userInfoURL string
	switch provider {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
	case "facebook":
		userInfoURL = "https://graph.facebook.com/me?fields=id,name,email"
	case "github":
		userInfoURL = "https://api.github.com/user"
	case "twitter":
		userInfoURL = "https://api.twitter.com/1.1/account/verify_credentials.json?include_email=true"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			s.logger.Warn("Failed to close response body", zap.Error(err))
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (s *service) createOrUpdateFirebaseUser(ctx context.Context, userInfo map[string]interface{}, provider string) (*models.FirebaseToken, error) {
	email, _ := userInfo["email"].(string)
	name, _ := userInfo["name"].(string)

	user, err := s.client.GetUserByEmail(ctx, email)
	if err != nil {
		params := (&firebaseauth.UserToCreate{}).
			Email(email).
			DisplayName(name)

		user, err = s.client.CreateUser(ctx, params)
		if err != nil {
			return nil, err
		}

		// 創建本地用戶
		localUser := &models.User{
			Username:    name,
			Email:       email,
			FirebaseUID: user.UID,
			Provider:    provider,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if _, err := s.userStore.CreateUser(ctx, localUser); err != nil {
			return nil, fmt.Errorf("failed to create local user: %w", err)
		}
	} else {
		// 更新用戶信息
		params := (&firebaseauth.UserToUpdate{}).
			DisplayName(name)

		if _, err = s.client.UpdateUser(ctx, user.UID, params); err != nil {
			s.logger.Error("Failed to update Firebase user", zap.Error(err))
			return nil, err
		}

		// 更新本地用戶
		localUser, err := s.userStore.FindUserByFirebaseUID(ctx, user.UID)
		if err != nil {
			return nil, fmt.Errorf("failed to find local user: %w", err)
		}
		localUser.Username = name
		localUser.Provider = provider
		localUser.UpdatedAt = time.Now()
		// if err := s.userStore.UpdateUser(ctx, localUser); err != nil {
		// 	return nil, fmt.Errorf("failed to update local user: %w", err)
		// }
	}

	token, err := s.client.CustomToken(ctx, user.UID)
	if err != nil {
		return nil, err
	}

	return &models.FirebaseToken{Token: token}, nil
}

// providerEndpoint returns the endpoint for the OAuth provider.
func providerEndpoint(provider string) oauth2.Endpoint {
	switch provider {
	case "google":
		return google.Endpoint
	case "facebook":
		return oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/v12.0/dialog/oauth",
			TokenURL: "https://graph.facebook.com/v12.0/oauth/access_token",
		}
	case "github":
		return oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		}
	case "twitter":
		return oauth2.Endpoint{
			AuthURL:  "https://api.twitter.com/oauth/authenticate",
			TokenURL: "https://api.twitter.com/oauth/access_token",
		}
	default:
		return oauth2.Endpoint{}
	}
}

// generateRandomState generates a random state string.
func generateRandomState() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)

	return fmt.Sprintf("%06d", n.Uint64())
}
