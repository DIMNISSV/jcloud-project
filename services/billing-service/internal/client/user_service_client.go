// services/billing-service/internal/client/user_service_client.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

//
// User Service Client
//

type UserDetails struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type UserServiceClient interface {
	GetUserDetails(ctx context.Context, userID int64) (*UserDetails, error)
}

type userServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewUserServiceClient() UserServiceClient {
	// The hostname is static because it's managed by Docker DNS
	return &userServiceClient{
		baseURL:    "http://user-service:8080",
		httpClient: &http.Client{},
	}
}

func (c *userServiceClient) GetUserDetails(ctx context.Context, userID int64) (*UserDetails, error) {
	endpoint := fmt.Sprintf("%s/internal/v1/users/%d", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user-service request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute user-service request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service returned non-200 status: %d", resp.StatusCode)
	}

	var details UserDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode user details response: %w", err)
	}

	return &details, nil
}
