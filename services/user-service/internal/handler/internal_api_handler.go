// services/user-service/internal/handler/internal_api_handler.go
package handler

import (
	"jcloud-project/user-service/internal/service"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type InternalApiHandler struct {
	service service.UserService
}

func NewInternalApiHandler(s service.UserService) *InternalApiHandler {
	return &InternalApiHandler{service: s}
}

// GetInternalUserDetails предоставляет детали для межсервисного взаимодействия.
func (h *InternalApiHandler) GetInternalUserDetails(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	user, err := h.service.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	// Возвращаем только необходимые поля для внутренних клиентов
	return c.JSON(http.StatusOK, echo.Map{
		"id":    user.ID,
		"email": user.Email,
	})
}
