package middleware

import (
	"errors"

	"github.com/labstack/echo/v4"
	"goflare.io/auth/authentication"
)

// AuthenticationMiddleware represents middleware for authentication
type AuthenticationMiddleware struct {
	authentication authentication.Service
}

// NewAuthenticationMiddleware creates a new instance of the AuthenticationMiddleware
// with the provided AuthenticationService and returns a pointer to it.
// It initializes the logger using the InfrastructurePackageName constant.
func NewAuthenticationMiddleware(
	authentication authentication.Service,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		authentication: authentication,
	}
}

// AuthorizeUser authorizes the user by checking the presence of a token in the request header.
// If the token is missing, it returns an error response with the error code ErrorCodeMissingToken and
// the message "missing token".
// If the token is present, it verifies the token using the VerifyPasetoToken method of the AuthenticationService.
// If the verification fails due to an invalid or expired token, it returns an error response with the
// error code ErrorCodeInvalidToken and the corresponding error.
// If the verification is successful, it checks if the role extracted from the token is either RoleCustomer or RoleAdmin.
// If not, it returns an error response with the error code ErrorCodeNoPermission.
// If the role is valid, it calls the next handler and passes the echo.Context object to it.
func (middleware *AuthenticationMiddleware) AuthorizeUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		tokenStr := c.Request().Header.Get("Authorization")
		if tokenStr == "" {
			return errors.New("missing token")
		}

		userID, err := middleware.authentication.ValidateToken(tokenStr)
		if err != nil {
			return err
		}

		c.Set("user_id", userID)

		return next(c)
	}
}
