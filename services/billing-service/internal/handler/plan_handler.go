// services/billing-service/internal/handler/plan_handler.go
package handler

import (
	"jcloud-project/billing-service/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type PlanHandler struct {
	service service.BillingService
}

func NewPlanHandler(s service.BillingService) *PlanHandler {
	return &PlanHandler{service: s}
}

func (h *PlanHandler) GetAllPlans(c echo.Context) error {
	plans, err := h.service.GetAllPlans(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, plans)
}
