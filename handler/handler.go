package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"goflare.io/auth"
	"goflare.io/auth/models/enum"
)

type UserHandler interface {
	Login(c echo.Context) error
	Register(c echo.Context) error
	CheckPermission(c echo.Context) error
}

type userHandler struct {
	Authentication auth.Authentication
}

func NewUserHandler(
	Authentication auth.Authentication,
) UserHandler {
	return &userHandler{
		Authentication: Authentication,
	}
}

type AuthHandler struct {
	AuthService auth.Authentication
}

type OAuthLoginRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google apple"`
	IDToken  string `json:"id_token" validate:"required"`
}

func (uh *userHandler) Login(c echo.Context) error {

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	paseto, err := uh.Authentication.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, paseto)
}

func (uh *userHandler) Register(c echo.Context) error {

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email or password is empty")
	}

	paseto, err := uh.Authentication.Register(c.Request().Context(), req.Username, req.Password, req.Email, req.Phone)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, paseto)
}

func (uh *userHandler) CheckPermission(c echo.Context) error {

	var req CheckPermissionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID := c.Get("user_id").(uint32)

	ok, err := uh.Authentication.CheckPermission(c.Request().Context(), userID, enum.ResourceType(req.Resource), enum.ActionType(req.Action))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ok)
}

func (uh *userHandler) OAuthLogin(c echo.Context) error {
	var req OAuthLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	token, err := uh.Authentication.OAuthLoginWithFirebase(c.Request().Context(), req.Provider, req.IDToken)
	if err != nil {
		return c.JSON(err.(*echo.HTTPError).Code, echo.Map{"error": err.(*echo.HTTPError).Message})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "OAuth login successful",
		"token":   token.Token,
	})
}
