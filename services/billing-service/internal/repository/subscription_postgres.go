// services/billing-service/internal/repository/subscription_postgres.go
package repository

import (
	"context"
	"errors"
	"jcloud-project/billing-service/internal/domain"
	"jcloud-project/libs/go-common/ierr"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type subscriptionPostgresRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionPostgresRepository(db *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionPostgresRepository{db: db}
}

func (r *subscriptionPostgresRepository) Create(ctx context.Context, userID, planID int64) error {
	query := `
		INSERT INTO user_subscriptions (user_id, plan_id, status, starts_at, ends_at)
		VALUES ($1, $2, 'ACTIVE', NOW(), NOW() + INTERVAL '100 year')`
	_, err := r.db.Exec(ctx, query, userID, planID)
	return err
}

func (r *subscriptionPostgresRepository) Update(ctx context.Context, userID, newPlanID int64, newEndDate time.Time) error {
	query := `
		UPDATE user_subscriptions 
		SET plan_id = $1, status = 'ACTIVE', starts_at = NOW(), ends_at = $2, updated_at = NOW()
		WHERE user_id = $3`
	_, err := r.db.Exec(ctx, query, newPlanID, newEndDate, userID)
	return err
}

func (r *subscriptionPostgresRepository) FindDetailsByUserID(ctx context.Context, userID int64) (*domain.UserSubscriptionDetails, error) {
	query := `
		SELECT p.name, s.status, s.ends_at FROM user_subscriptions s
		JOIN subscription_plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status = 'ACTIVE'`
	var d domain.UserSubscriptionDetails
	err := r.db.QueryRow(ctx, query, userID).Scan(&d.PlanName, &d.Status, &d.EndsAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ierr.ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *subscriptionPostgresRepository) FindPermissionsByUserID(ctx context.Context, userID int64) (*domain.SubscriptionPlan, error) {
	query := `
		SELECT p.permissions FROM user_subscriptions s
		JOIN subscription_plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status = 'ACTIVE'`
	var p domain.SubscriptionPlan
	err := r.db.QueryRow(ctx, query, userID).Scan(&p.Permissions)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ierr.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}
