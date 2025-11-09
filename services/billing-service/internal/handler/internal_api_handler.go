// services/billing-service/internal/handler/internal_api_handler.go
package handler

import (
	"jcloud-project/billing-service/internal/service"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type InternalApiHandler struct {
	service service.BillingService
}

func NewInternalApiHandler(s service.BillingService) *InternalApiHandler {
	return &InternalApiHandler{service: s}
}

func (h *InternalApiHandler) GetUserPermissions(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	permissions, err := h.service.GetUserPermissions(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, permissions)
}

type createSubscriptionRequest struct {
	UserID   int64  `json:"userId"`
	PlanName string `json:"planName"`
}

func (h *InternalApiHandler) CreateSubscription(c echo.Context) error {
	var req createSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}

	err := h.service.CreateInitialSubscription(c.Request().Context(), req.UserID, req.PlanName)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "subscription created successfully"})
}
