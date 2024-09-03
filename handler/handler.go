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

	paseto, err := uh.Authentication.Register(c.Request().Context(), req.Username, req.Password, req.Email)
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
