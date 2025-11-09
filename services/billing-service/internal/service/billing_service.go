// services/billing-service/internal/service/billing_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"jcloud-project/billing-service/internal/client"
	"jcloud-project/billing-service/internal/domain"
	"jcloud-project/billing-service/internal/repository"
	"jcloud-project/libs/go-common/ierr"
	"log"
	"time"
)

type BillingService interface {
	GetUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error)
	CreateInitialSubscription(ctx context.Context, userID int64, planName string) error
	GetAllPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	GetUserSubscription(ctx context.Context, userID int64) (*domain.UserSubscriptionDetails, error)
	ChangeSubscription(ctx context.Context, userID, newPlanID int64) error
}

type billingService struct {
	planRepo        repository.PlanRepository
	subRepo         repository.SubscriptionRepository
	nextcloudClient client.NextcloudClient
	userSvcClient   client.UserServiceClient
}

func NewBillingService(planRepo repository.PlanRepository, subRepo repository.SubscriptionRepository, ncClient client.NextcloudClient, userSvcClient client.UserServiceClient) BillingService {
	return &billingService{
		planRepo:        planRepo,
		subRepo:         subRepo,
		nextcloudClient: ncClient,
		userSvcClient:   userSvcClient,
	}
}

func (s *billingService) GetUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error) {
	plan, err := s.subRepo.FindPermissionsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ierr.ErrNotFound) {
			return make(map[string]interface{}), nil // No subscription = empty permissions
		}
		return nil, err
	}
	return plan.Permissions, nil
}

func (s *billingService) CreateInitialSubscription(ctx context.Context, userID int64, planName string) error {
	plan, err := s.planRepo.FindByName(ctx, planName)
	if err != nil {
		return fmt.Errorf("could not find plan '%s': %w", planName, err)
	}
	return s.subRepo.Create(ctx, userID, plan.ID)
}

func (s *billingService) GetAllPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	return s.planRepo.FindAllActive(ctx)
}

func (s *billingService) GetUserSubscription(ctx context.Context, userID int64) (*domain.UserSubscriptionDetails, error) {
	return s.subRepo.FindDetailsByUserID(ctx, userID)
}

func (s *billingService) ChangeSubscription(ctx context.Context, userID, newPlanID int64) error {
	plan, err := s.planRepo.FindByID(ctx, newPlanID)
	if err != nil {
		return fmt.Errorf("plan not found: %w", ierr.ErrNotFound)
	}
	if !plan.IsActive {
		return fmt.Errorf("cannot switch to an inactive plan: %w", ierr.ErrConflict)
	}

	var newEndDate time.Time
	if plan.Name == "Free" {
		newEndDate = time.Now().AddDate(100, 0, 0)
	} else {
		newEndDate = time.Now().AddDate(0, 1, 0)
	}

	if err := s.subRepo.Update(ctx, userID, newPlanID, newEndDate); err != nil {
		return err
	}

	go s.syncUserQuotaWithNextcloud(userID, plan.Permissions)

	log.Printf("User %d successfully changed subscription to plan %d. Quota sync initiated.", userID, newPlanID)
	return nil
}

func (s *billingService) syncUserQuotaWithNextcloud(userID int64, permissions map[string]interface{}) {
	ctx := context.Background()

	userDetails, err := s.userSvcClient.GetUserDetails(ctx, userID)
	if err != nil {
		log.Printf("CRITICAL: Failed to get details for user %d to sync quota: %v", userID, err)
		return
	}

	quotaGB, ok := permissions["storage_quota_gb"].(float64)
	if !ok {
		log.Printf("Warning: 'storage_quota_gb' not found for user %d", userID)
		return
	}

	if err := s.nextcloudClient.SetUserQuota(ctx, userDetails.Email, int(quotaGB)); err != nil {
		log.Printf("CRITICAL: Failed to set Nextcloud quota for user %s (ID: %d): %v", userDetails.Email, userID, err)
		return
	}

	log.Printf("Successfully synced quota for user %s to %d GB.", userDetails.Email, int(quotaGB))
}
