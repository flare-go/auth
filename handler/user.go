package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"goflare.io/auth/authentication"
	"goflare.io/auth/firebase"
	"goflare.io/auth/models/enum"
)

// UserHandler handles user-related endpoints.
type UserHandler struct {
	authentication  authentication.Service
	firebaseService firebase.Service
	logger          *zap.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(
	authentication authentication.Service,
	firebaseService firebase.Service,
	logger *zap.Logger,
) *UserHandler {
	return &UserHandler{
		authentication:  authentication,
		firebaseService: firebaseService,
		logger:          logger,
	}
}

// Login handles the login endpoint.
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uh.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := uh.authentication.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		uh.logger.Error("Failed to login user", zap.Error(err))
		http.Error(w, "Login failed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(token); err != nil {
		return
	}
}

// Register handles the user registration endpoint. It decodes the request body, registers the user, and returns a token.
func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uh.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := uh.authentication.Register(r.Context(), req.Username, req.Password, req.Email, "")
	if err != nil {
		uh.logger.Error("Failed to register user", zap.Error(err))
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(token); err != nil {
		return
	}
}

// CheckPermission handles the check permission endpoint.
func (uh *UserHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	var req CheckPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uint64)

	if _, err := uh.authentication.CheckPermission(r.Context(), userID, enum.ResourceType(req.Resource), enum.ActionType(req.Action)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// FirebaseLogin handles the Firebase login endpoint.
func (uh *UserHandler) FirebaseLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken  string `json:"id_token"`
		Provider string `json:"provider"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uh.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := uh.firebaseService.Login(r.Context(), req.IDToken, req.Provider)
	if err != nil {
		uh.logger.Error("Failed to login with Firebase", zap.Error(err))
		http.Error(w, "Firebase login failed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(token); err != nil {
		return
	}
}

// RegisterWithFirebase FirebaseRegister handles the Firebase registration endpoint. It decodes the request body, registers the user, and returns a token.
func (uh *UserHandler) RegisterWithFirebase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uh.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := uh.firebaseService.Register(r.Context(), req.Username, req.Password, req.Email, req.Phone)
	if err != nil {
		uh.logger.Error("Failed to register with Firebase", zap.Error(err))
		http.Error(w, "Firebase registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(token); err != nil {
		return
	}
}

// OAuthLogin handles the OAuth login endpoint.
func (uh *UserHandler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider string `json:"provider"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uh.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	authURL, err := uh.firebaseService.GetOAuthURL(req.Provider)
	if err != nil {
		uh.logger.Error("Failed to get OAuth URL", zap.Error(err))
		http.Error(w, "Failed to initiate OAuth flow", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// OAuthCallback handles the OAuth callback endpoint.
func (uh *UserHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		uh.logger.Error("Missing OAuth code")
		http.Error(w, "Missing OAuth code", http.StatusBadRequest)
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		uh.logger.Error("Missing OAuth provider")
		http.Error(w, "Missing OAuth provider", http.StatusBadRequest)
		return
	}

	token, err := uh.firebaseService.ExchangeOAuthToken(r.Context(), code, provider)
	if err != nil {
		uh.logger.Error("無法交換 OAuth 代碼", zap.Error(err))
		http.Error(w, "無法完成 OAuth 流程", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(token); err != nil {
		return
	}
}
