// internal/service/user_service.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"jcloud-project/user-service/internal/domain"
	"jcloud-project/user-service/internal/repository"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

//
// JWT Custom Claims
//

type JwtCustomClaims struct {
	UserID      int64                  `json:"user_id"`
	Permissions map[string]interface{} `json:"perms,omitempty"`
	jwt.RegisteredClaims
}

//
// User Service Interface
//

type UserService interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, userID int64) (*domain.User, error)
}

//
// Service Implementation
//

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	// --- Assign Default Subscription ---
	if err := s.assignDefaultSubscription(ctx, user.ID); err != nil {
		log.Printf("CRITICAL: Failed to assign default subscription for new user %d: %v", user.ID, err)
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Passwords don't match
		return "", errors.New("invalid credentials")
	}

	//
	// Generate JWT
	//

	permissions, err := s.fetchUserPermissions(ctx, user.ID)
	if err != nil {
		// Log the error but don't fail the login. Proceed with empty permissions.
		log.Printf("Warning: Could not fetch permissions for user %d: %v", user.ID, err)
		permissions = make(map[string]interface{})
	}

	// Set custom claims
	claims := &JwtCustomClaims{
		UserID:      user.ID,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	jwtSecret := os.Getenv("JWT_SECRET")
	t, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	return s.repo.FindByID(ctx, userID)
}

func (s *userService) fetchUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error) {
	// The hostname "billing-service" is resolved by Docker's internal DNS.
	billingURL := fmt.Sprintf("http://billing-service:8082/internal/v1/permissions/%d", userID)

	req, err := http.NewRequestWithContext(ctx, "GET", billingURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call billing service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("billing service returned status %d", resp.StatusCode)
	}

	var permissions map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&permissions); err != nil {
		return nil, fmt.Errorf("failed to decode permissions response: %w", err)
	}

	return permissions, nil
}

func (s *userService) assignDefaultSubscription(ctx context.Context, userID int64) error {
	billingURL := "http://billing-service:8082/internal/v1/subscriptions"

	reqBody, err := json.Marshal(map[string]interface{}{
		"userId":   userID,
		"planName": "Free",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", billingURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call billing service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("billing service returned non-201 status: %d", resp.StatusCode)
	}

	return nil
}
