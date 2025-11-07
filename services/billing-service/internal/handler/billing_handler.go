// internal/handler/billing_handler.go
package handler

import (
	"jcloud-project/billing-service/internal/service"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type BillingHandler struct {
	service service.BillingService
}

func NewBillingHandler(s service.BillingService) *BillingHandler {
	return &BillingHandler{service: s}
}

func (h *BillingHandler) GetUserPermissions(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	permissions, err := h.service.GetUserPermissions(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not retrieve permissions"})
	}

	return c.JSON(http.StatusOK, permissions)
}

type createSubscriptionRequest struct {
	UserID   int64  `json:"userId"`
	PlanName string `json:"planName"`
}

func (h *BillingHandler) CreateSubscription(c echo.Context) error {
	var req createSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}

	if req.UserID == 0 || req.PlanName == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "userId and planName are required"})
	}

	err := h.service.CreateInitialSubscription(c.Request().Context(), req.UserID, req.PlanName)
	if err != nil {
		// In a real app, check for specific errors like "plan not found" or "user already subscribed"
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not create subscription"})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "subscription created successfully"})
}
