// services/billing-service/internal/repository/interfaces.go
package repository

import (
	"context"
	"jcloud-project/billing-service/internal/domain"
	"time"
)

type PlanRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.SubscriptionPlan, error)
	FindByName(ctx context.Context, name string) (*domain.SubscriptionPlan, error)
	FindAllActive(ctx context.Context) ([]domain.SubscriptionPlan, error)
}

type SubscriptionRepository interface {
	Create(ctx context.Context, userID, planID int64) error
	Update(ctx context.Context, userID, newPlanID int64, newEndDate time.Time) error
	FindDetailsByUserID(ctx context.Context, userID int64) (*domain.UserSubscriptionDetails, error)
	FindPermissionsByUserID(ctx context.Context, userID int64) (*domain.SubscriptionPlan, error)
}
