// internal/service/billing_service.go
package service

import (
	"context"
	"errors"
	"jcloud-project/billing-service/internal/domain"
	"jcloud-project/billing-service/internal/repository"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type BillingService interface {
	GetUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error)
	CreateInitialSubscription(ctx context.Context, userID int64, planName string) error
	GetAllPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	GetUserSubscription(ctx context.Context, userID int64) (*domain.UserSubscriptionDetails, error)
	ChangeSubscription(ctx context.Context, userID, newPlanID int64) error
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

func (s *billingService) GetUserSubscription(ctx context.Context, userID int64) (*domain.UserSubscriptionDetails, error) {
	details, err := s.repo.FindSubscriptionDetailsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return a specific, recognizable error if the user has no subscription
			return nil, errors.New("subscription not found")
		}
		return nil, err
	}
	return details, nil
}

func (s *billingService) ChangeSubscription(ctx context.Context, userID, newPlanID int64) error {
	// Step 1: Validate that the new plan exists and is active
	plan, err := s.repo.FindPlanByID(ctx, newPlanID)
	if err != nil {
		return err // "plan not found" or other DB error
	}
	if !plan.IsActive {
		return errors.New("cannot switch to an inactive plan")
	}

	// Step 2: Determine the new expiration date based on the plan
	var newEndDate time.Time
	if plan.Name == "Free" {
		newEndDate = time.Now().AddDate(100, 0, 0) // 100 years for Free plan
	} else {
		// For paid plans, set it to 30 days from now.
		// In a real app, this would be tied to a successful payment.
		newEndDate = time.Now().AddDate(0, 1, 0) // 1 month
	}

	// Step 3: Update the user's subscription in the database
	err = s.repo.UpdateUserSubscription(ctx, userID, newPlanID, newEndDate)
	if err != nil {
		return err
	}

	// Step 4 (Future): Trigger side-effects, like updating Nextcloud quota.
	// We will implement this in the next epic.
	log.Printf("User %d successfully changed subscription to plan %d (%s)", userID, newPlanID, plan.Name)

	return nil
}
