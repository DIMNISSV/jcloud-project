// services/user-service/internal/service/user_service.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"jcloud-project/libs/go-common/ierr"
	commontypes "jcloud-project/libs/go-common/types/jwt"
	"jcloud-project/user-service/internal/domain"
	"jcloud-project/user-service/internal/repository"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, userID int64) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.UserPublic, error)
	PatchUser(ctx context.Context, userID int64, email *string, role *string) (*domain.User, error)
}

type userService struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewUserService(repo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *userService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters: %w", ierr.ErrConflict)
	}
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

	if err := s.assignDefaultSubscription(ctx, user.ID); err != nil {
		log.Printf("CRITICAL: Failed to assign default subscription for new user %d: %v", user.ID, err)
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		// Неважно, не найден юзер или другая ошибка бд, для безопасности возвращаем одну ошибку
		return "", ierr.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Пароли не совпадают
		return "", ierr.ErrInvalidCredentials
	}

	permissions, err := s.fetchUserPermissions(ctx, user.ID)
	if err != nil {
		log.Printf("Warning: Could not fetch permissions for user %d: %v", user.ID, err)
		permissions = make(map[string]interface{})
	}

	claims := &commontypes.JwtCustomClaims{
		UserID:      user.ID,
		Role:        user.Role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	return s.repo.FindByID(ctx, userID)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.UserPublic, error) {
	return s.repo.FindAll(ctx)
}

func (s *userService) PatchUser(ctx context.Context, userID int64, email *string, role *string) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err // repo.FindByID уже возвращает ierr.ErrNotFound
	}

	if email != nil {
		user.Email = *email
	}
	if role != nil {
		user.Role = *role
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// --- Private methods ---

func (s *userService) fetchUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error) {
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
