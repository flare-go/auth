package middleware

import (
	"context"
	"net/http"

	"goflare.io/auth/authentication"
	"goflare.io/auth/models/enum"
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

		ctx := context.WithValue(r.Context(), enum.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
