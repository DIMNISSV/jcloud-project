// services/billing-service/internal/repository/plan_postgres.go
package repository

import (
	"context"
	"errors"
	"jcloud-project/billing-service/internal/domain"
	"jcloud-project/libs/go-common/ierr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type planPostgresRepository struct {
	db *pgxpool.Pool
}

func NewPlanPostgresRepository(db *pgxpool.Pool) PlanRepository {
	return &planPostgresRepository{db: db}
}

func (r *planPostgresRepository) FindByID(ctx context.Context, id int64) (*domain.SubscriptionPlan, error) {
	query := `SELECT id, name, price, permissions, is_active FROM subscription_plans WHERE id = $1`
	var p domain.SubscriptionPlan
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Permissions, &p.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ierr.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *planPostgresRepository) FindByName(ctx context.Context, name string) (*domain.SubscriptionPlan, error) {
	query := `SELECT id, name, price, permissions, is_active FROM subscription_plans WHERE name = $1`
	var p domain.SubscriptionPlan
	err := r.db.QueryRow(ctx, query, name).Scan(&p.ID, &p.Name, &p.Price, &p.Permissions, &p.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ierr.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *planPostgresRepository) FindAllActive(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	query := `SELECT id, name, price, permissions, is_active FROM subscription_plans WHERE is_active = true ORDER BY price ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []domain.SubscriptionPlan
	for rows.Next() {
		var p domain.SubscriptionPlan
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Permissions, &p.IsActive); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, nil
}
