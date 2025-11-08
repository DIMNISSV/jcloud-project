// internal/service/billing_service.go
package service

import (
	"context"
	"errors"
	"jcloud-project/billing-service/internal/domain"
	"jcloud-project/billing-service/internal/repository"

	"github.com/jackc/pgx/v5"
)

type BillingService interface {
	GetUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error)
	CreateInitialSubscription(ctx context.Context, userID int64, planName string) error
	GetAllPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
}

type billingService struct {
	repo repository.BillingRepository
}

func NewBillingService(repo repository.BillingRepository) BillingService {
	return &billingService{repo: repo}
}

func (s *billingService) GetUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error) {
	plan, err := s.repo.FindPermissionsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If user has no subscription, return empty permissions.
			// In a real app, you might assign a default 'Free' plan here.
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	return plan.Permissions, nil
}

func (s *billingService) CreateInitialSubscription(ctx context.Context, userID int64, planName string) error {
	return s.repo.CreateInitialSubscription(ctx, userID, planName)
}

func (s *billingService) GetAllPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	return s.repo.FindAllActivePlans(ctx)
}
