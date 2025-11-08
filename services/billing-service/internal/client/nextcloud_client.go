// services/billing-service/internal/client/nextcloud_client.go
package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

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

// SetUserQuota updates a user's storage quota in Nextcloud.
func (c *nextcloudClient) SetUserQuota(ctx context.Context, username string, quotaGB int) error {
	// Using v1.php as per the working curl example.
	// The username is URL-encoded to handle special characters like '@'.
	endpoint := fmt.Sprintf("%s/ocs/v1.php/cloud/users/%s", c.baseURL, url.PathEscape(username))

	data := url.Values{}
	data.Set("key", "quota")
	data.Set("value", fmt.Sprintf("%d GB", quotaGB))

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create nextcloud request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("OCS-APIRequest", "true")
	req.SetBasicAuth(c.apiUser, c.apiPassword)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute nextcloud request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("nextcloud API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
