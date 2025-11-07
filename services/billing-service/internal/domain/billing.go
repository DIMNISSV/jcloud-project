// internal/domain/billing.go
package domain

import "time"

// SubscriptionPlan is the template for a subscription (e.g., "Free", "Pro")
type SubscriptionPlan struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"` // Monthly price
	// Permissions holds all features and limits for this plan.
	// Using map[string]interface{} for flexibility, stored as JSONB in Postgres.
	Permissions map[string]interface{} `json:"permissions"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// UserSubscription is an instance of a user subscribed to a specific plan.
type UserSubscription struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	PlanID   int64     `json:"plan_id"`
	Status   string    `json:"status"` // e.g., "ACTIVE", "CANCELED", "PAST_DUE"
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}
