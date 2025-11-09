// services/user-service/internal/handler/admin_handler.go
package handler

import (
	commontypes "jcloud-project/libs/go-common/types/jwt"
	"jcloud-project/user-service/internal/service"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	service service.UserService
}

func NewAdminHandler(s service.UserService) *AdminHandler {
	return &AdminHandler{service: s}
}

func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userToken, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
		}
		claims, ok := userToken.Claims.(*commontypes.JwtCustomClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
		}
		if claims.Role != "ADMIN" {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "access forbidden: administrator role required"})
		}
		return next(c)
	}
}

func (h *AdminHandler) GetAllUsers(c echo.Context) error {
	users, err := h.service.GetAllUsers(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, users)
}

type patchUserRequest struct {
	Email *string `json:"email,omitempty"`
	Role  *string `json:"role,omitempty"`
}

func (h *AdminHandler) PatchUser(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	var req patchUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}

	user, err := h.service.PatchUser(c.Request().Context(), userID, req.Email, req.Role)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}
