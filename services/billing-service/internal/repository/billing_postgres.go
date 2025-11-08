// internal/repository/billing_postgres.go
package repository

import (
	"context"
	"encoding/json"
	"jcloud-project/billing-service/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingRepository interface {
	FindPermissionsByUserID(ctx context.Context, userID int64) (*domain.SubscriptionPlan, error)
	CreateInitialSubscription(ctx context.Context, userID int64, planName string) error
	FindAllActivePlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	FindSubscriptionDetailsByUserID(ctx context.Context, userID int64) (*domain.UserSubscriptionDetails, error)
}

type billingPostgresRepository struct {
	db *pgxpool.Pool
}

func NewBillingPostgresRepository(db *pgxpool.Pool) BillingRepository {
	return &billingPostgresRepository{db: db}
}

func (r *billingPostgresRepository) FindPermissionsByUserID(ctx context.Context, userID int64) (*domain.SubscriptionPlan, error) {
	query := `
        SELECT p.permissions
        FROM subscription_plans p
        JOIN user_subscriptions us ON p.id = us.plan_id
        WHERE us.user_id = $1 AND us.status = 'ACTIVE'
    `
	var permissionsJSON []byte
	err := r.db.QueryRow(ctx, query, userID).Scan(&permissionsJSON)
	if err != nil {
		return nil, err // This will correctly handle pgx.ErrNoRows
	}

	var plan domain.SubscriptionPlan
	plan.Permissions = make(map[string]interface{})
	if err := json.Unmarshal(permissionsJSON, &plan.Permissions); err != nil {
		return nil, err
	}

	return &plan, nil
}

func (r *billingPostgresRepository) CreateInitialSubscription(ctx context.Context, userID int64, planName string) error {
	// In a transaction, first get the plan ID, then create the subscription
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Rollback on error

	var planID int64
	err = tx.QueryRow(ctx, "SELECT id FROM subscription_plans WHERE name = $1", planName).Scan(&planID)
	if err != nil {
		return err // Plan not found
	}

	// For a "Free" plan, we can set a very long expiry date
	endsAt := time.Now().AddDate(100, 0, 0) // 100 years from now

	_, err = tx.Exec(ctx,
		"INSERT INTO user_subscriptions (user_id, plan_id, status, ends_at) VALUES ($1, $2, 'ACTIVE', $3)",
		userID, planID, endsAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx) // Commit the transaction
}

// FindAllActivePlans retrieves all active subscription plans from the database.
func (r *billingPostgresRepository) FindAllActivePlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	// NOTE: We select `is_active` here so that RowToStructByName can map it.
	query := `
        SELECT id, name, price, permissions, is_active, created_at, updated_at
        FROM subscription_plans
        WHERE is_active = TRUE
        ORDER BY price ASC
    `
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// The correct, idiomatic way to use CollectRows with structs.
	// pgx.RowToStructByName handles mapping snake_case columns to PascalCase fields.
	// It also handles JSONB to map[string]interface{} automatically.
	plans, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.SubscriptionPlan])
	if err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *billingPostgresRepository) FindSubscriptionDetailsByUserID(ctx context.Context, userID int64) (*domain.UserSubscriptionDetails, error) {
	query := `
        SELECT p.name, us.status, us.ends_at
        FROM user_subscriptions us
        JOIN subscription_plans p ON us.plan_id = p.id
        WHERE us.user_id = $1
    `
	details := &domain.UserSubscriptionDetails{}
	err := r.db.QueryRow(ctx, query, userID).Scan(&details.PlanName, &details.Status, &details.EndsAt)
	if err != nil {
		return nil, err
	}
	return details, nil
}
