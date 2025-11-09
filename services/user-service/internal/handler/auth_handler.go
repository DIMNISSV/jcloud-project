// services/user-service/internal/handler/auth_handler.go
package handler

import (
	"jcloud-project/user-service/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service service.UserService
}

func NewAuthHandler(s service.UserService) *AuthHandler {
	return &AuthHandler{service: s}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}

	user, err := h.service.Register(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err // Передаем ошибку в центральный обработчик
	}

	return c.JSON(http.StatusCreated, user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}

	token, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err // Передаем ошибку в центральный обработчик
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}
