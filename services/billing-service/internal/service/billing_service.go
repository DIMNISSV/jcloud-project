// internal/service/billing_service.go
package service

import (
	"context"
	"errors"
	"log"
	"time"

	"jcloud-project/billing-service/internal/client"
	"jcloud-project/billing-service/internal/domain"
	"jcloud-project/billing-service/internal/repository"

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
	repo            repository.BillingRepository
	nextcloudClient client.NextcloudClient
	userSvcClient   client.UserServiceClient
}

func NewBillingService(repo repository.BillingRepository, ncClient client.NextcloudClient, userSvcClient client.UserServiceClient) BillingService {
	return &billingService{
		repo:            repo,
		nextcloudClient: ncClient,
		userSvcClient:   userSvcClient,
	}
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

	// Step 3: Update subscription in DB (код без изменений)
	err = s.repo.UpdateUserSubscription(ctx, userID, newPlanID, newEndDate)
	if err != nil {
		return err
	}

	// --- Step 4: Trigger Side-Effects ---
	// This is the new logic we are adding.
	go s.syncUserQuotaWithNextcloud(userID, plan.Permissions)

	log.Printf("User %d successfully changed subscription to plan %d (%s). Quota sync initiated.", userID, newPlanID, plan.Name)

	return nil
}

func (s *billingService) syncUserQuotaWithNextcloud(userID int64, permissions map[string]interface{}) {
	// We use a background context because the original request might have already finished.
	ctx := context.Background()

	// 4.1: Get user email from user-service
	userDetails, err := s.userSvcClient.GetUserDetails(ctx, userID)
	if err != nil {
		log.Printf("CRITICAL: Failed to get details for user %d to sync quota: %v", userID, err)
		return
	}

	// 4.2: Extract quota from permissions
	quotaGB, ok := permissions["storage_quota_gb"].(float64) // JSON numbers are float64
	if !ok {
		log.Printf("Warning: 'storage_quota_gb' not found or invalid in permissions for user %d", userID)
		return
	}

	// 4.3: Call Nextcloud API
	err = s.nextcloudClient.SetUserQuota(ctx, userDetails.Email, int(quotaGB))
	if err != nil {
		log.Printf("CRITICAL: Failed to set Nextcloud quota for user %s (ID: %d): %v", userDetails.Email, userID, err)
		return
	}

	log.Printf("Successfully synced quota for user %s to %d GB.", userDetails.Email, int(quotaGB))
}
