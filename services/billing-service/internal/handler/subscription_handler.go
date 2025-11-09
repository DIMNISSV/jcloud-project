// services/billing-service/internal/handler/subscription_handler.go
package handler

import (
	"jcloud-project/billing-service/internal/service"
	commontypes "jcloud-project/libs/go-common/types/jwt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type SubscriptionHandler struct {
	service service.BillingService
}

func NewSubscriptionHandler(s service.BillingService) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

func (h *SubscriptionHandler) GetUserSubscription(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*commontypes.JwtCustomClaims)

	subscription, err := h.service.GetUserSubscription(c.Request().Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, subscription)
}

type changeSubscriptionRequest struct {
	PlanID int64 `json:"planId"`
}

func (h *SubscriptionHandler) ChangeSubscription(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*commontypes.JwtCustomClaims)

	var req changeSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}

	err := h.service.ChangeSubscription(c.Request().Context(), claims.UserID, req.PlanID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "subscription updated successfully"})
}
