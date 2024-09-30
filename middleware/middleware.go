package middleware

import (
	"context"
	"net/http"

	"goflare.io/auth/authentication"
)

// AuthenticationMiddleware represents middleware for authentication
type AuthenticationMiddleware struct {
	authentication authentication.Service
}

// NewAuthenticationMiddleware creates a new instance of the AuthenticationMiddleware
func NewAuthenticationMiddleware(
	authentication authentication.Service,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		authentication: authentication,
	}
}

// AuthorizeUser is a middleware function that authorizes the user
func (middleware *AuthenticationMiddleware) AuthorizeUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		userID, err := middleware.authentication.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
