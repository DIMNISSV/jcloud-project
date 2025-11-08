// services/billing-service/internal/client/nextcloud_client.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

//
// OCSResponse defines the structure for a typical Nextcloud OCS JSON response.
// This allows us to properly parse the response and check the application-level status code.
//

type OCSResponse struct {
	Ocs OCS `json:"ocs"`
}

type OCS struct {
	Meta OCSMeta     `json:"meta"`
	Data interface{} `json:"data"` // Data can be of any type, so we use interface{}
}

type OCSMeta struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statuscode"`
	Message    string `json:"message"`
}

//
// Nextcloud Client
//

type NextcloudClient interface {
	SetUserQuota(ctx context.Context, username string, quotaGB int) error
}

type nextcloudClient struct {
	baseURL     string
	apiUser     string
	apiPassword string
	httpClient  *http.Client
}

func NewNextcloudClient(baseURL, apiUser, apiPassword string) NextcloudClient {
	return &nextcloudClient{
		baseURL:     baseURL,
		apiUser:     apiUser,
		apiPassword: apiPassword,
		httpClient:  &http.Client{},
	}
}

// SetUserQuota updates a user's storage quota in Nextcloud using the v2 JSON API.
func (c *nextcloudClient) SetUserQuota(ctx context.Context, username string, quotaGB int) error {
	// Use the v2 API endpoint to get JSON responses
	endpoint := fmt.Sprintf("%s/ocs/v2.php/cloud/users/%s", c.baseURL, url.PathEscape(username))

	// The request body remains form-urlencoded as it's a simple key-value update
	data := url.Values{}
	data.Set("key", "quota")
	data.Set("value", fmt.Sprintf("%d GB", quotaGB))

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create nextcloud request: %w", err)
	}

	// Set required headers for OCS API v2
	req.SetBasicAuth(c.apiUser, c.apiPassword)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("OCS-APIRequest", "true")
	req.Header.Add("Accept", "application/json") // <-- Crucial for getting a JSON response

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute nextcloud request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("nextcloud API returned non-200 HTTP status: %d", resp.StatusCode)
	}

	// Decode the JSON response
	var ocsResponse OCSResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocsResponse); err != nil {
		return fmt.Errorf("failed to decode nextcloud JSON response: %w", err)
	}

	// Check the application-level status code inside the JSON response
	if ocsResponse.Ocs.Meta.StatusCode != 200 {
		return fmt.Errorf("nextcloud OCS API returned an error: status=%d, message='%s'",
			ocsResponse.Ocs.Meta.StatusCode, ocsResponse.Ocs.Meta.Message)
	}

	return nil
}
