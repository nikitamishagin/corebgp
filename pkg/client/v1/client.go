package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"time"

	"github.com/nikitamishagin/corebgp/internal/model"
)

// APIClient represents the client for interacting with the API server.
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new API client instance.
func NewAPIClient(baseURL *string, timeout time.Duration) *APIClient {
	return &APIClient{
		baseURL: *baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// HealthCheck checks the health status of the API server (Version 1).
func (c *APIClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/healthz", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status code %d", resp.StatusCode)
	}

	return nil
}

// ListAnnouncements returns a list of announcement paths (IDs) globally.
func (c *APIClient) ListAnnouncements(ctx context.Context) (model.APIResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/announcements/", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return model.APIResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.APIResponse{}, err
	}
	defer resp.Body.Close()

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return apiResponse, fmt.Errorf("failed to list announcements: status code %d", resp.StatusCode)
	}

	return apiResponse, nil
}

// GetAllAnnouncements retrieves all announcements globally.
func (c *APIClient) GetAllAnnouncements(ctx context.Context) (model.APIResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/announcements/all", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return model.APIResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.APIResponse{}, err
	}
	defer resp.Body.Close()

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return apiResponse, fmt.Errorf("failed to fetch all announcements: status code %d", resp.StatusCode)
	}

	return apiResponse, nil
}

// ListAnnouncementsByProject retrieves announcement paths for a given project.
func (c *APIClient) ListAnnouncementsByProject(ctx context.Context, project string) (model.APIResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/announcements/%s/", c.baseURL, project)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return model.APIResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.APIResponse{}, err
	}
	defer resp.Body.Close()

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return apiResponse, fmt.Errorf("failed to list announcements for project %s: status code %d", project, resp.StatusCode)
	}

	return apiResponse, nil
}

// GetAllAnnouncementsByProject retrieves all announcements for a project.
func (c *APIClient) GetAllAnnouncementsByProject(ctx context.Context, project string) (model.APIResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/announcements/%s/all", c.baseURL, project)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return model.APIResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.APIResponse{}, err
	}
	defer resp.Body.Close()

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return apiResponse, fmt.Errorf("failed to fetch announcements for project %s: status code %d", project, resp.StatusCode)
	}

	return apiResponse, nil
}

// GetAnnouncement retrieves a specific announcement.
func (c *APIClient) GetAnnouncement(ctx context.Context, project, name string) (model.APIResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/announcements/%s/%s", c.baseURL, project, name)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return model.APIResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.APIResponse{}, err
	}
	defer resp.Body.Close()

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return apiResponse, fmt.Errorf("announcement not found")
	}

	if resp.StatusCode != http.StatusOK {
		return apiResponse, fmt.Errorf("failed to fetch announcement: status code %d", resp.StatusCode)
	}

	return apiResponse, nil
}

// CreateAnnouncement creates a new announcement.
func (c *APIClient) CreateAnnouncement(ctx context.Context, announcement *model.Announcement) (model.APIResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/announcements/%s/%s", c.baseURL, announcement.Meta.Project, announcement.Meta.Name)

	data, err := json.Marshal(announcement)
	if err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to marshal announcement: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, bytes.NewBuffer(data))
	if err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode == http.StatusConflict {
		return apiResponse, fmt.Errorf("announcement already exists")
	}

	if resp.StatusCode != http.StatusCreated {
		return apiResponse, fmt.Errorf("failed to create announcement: status code %d", resp.StatusCode)
	}

	return apiResponse, nil
}

// UpdateAnnouncement updates an existing announcement.
func (c *APIClient) UpdateAnnouncement(ctx context.Context, announcement *model.Announcement) (model.APIResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/announcements/%s/%s", c.baseURL, announcement.Meta.Project, announcement.Meta.Name)

	data, err := json.Marshal(announcement)
	if err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to marshal announcement: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", baseURL, bytes.NewBuffer(data))
	if err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return apiResponse, fmt.Errorf("announcement not found")
	}

	if resp.StatusCode != http.StatusOK {
		return apiResponse, fmt.Errorf("failed to update announcement: status code %d", resp.StatusCode)
	}

	return apiResponse, nil
}

// DeleteAnnouncement deletes an announcement by project and name.
func (c *APIClient) DeleteAnnouncement(ctx context.Context, project, name string) (model.APIResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/announcements/%s/%s", c.baseURL, project, name)

	req, err := http.NewRequestWithContext(ctx, "DELETE", baseURL, nil)
	if err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.APIResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return apiResponse, fmt.Errorf("announcement not found")
	}

	if resp.StatusCode != http.StatusOK {
		return apiResponse, fmt.Errorf("failed to delete announcement: status code %d", resp.StatusCode)
	}

	return apiResponse, nil
}

// WatchAnnouncements establishes a WebSocket connection to watch announcements.
func (c *APIClient) WatchAnnouncements(ctx context.Context, onEvent func(event model.Event)) error {
	parsedURL, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	// Replace 'http' with 'ws' and 'https' with 'wss'
	switch parsedURL.Scheme {
	case "http":
		parsedURL.Scheme = "ws"
	case "https":
		parsedURL.Scheme = "wss"
	default:
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Append the path for WebSocket announcements
	parsedURL.Path = "/v1/watch/announcements/"

	// Build the WebSocket URL
	webSocketURL := parsedURL.String()

	// Initialize WebSocket connection
	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, webSocketURL, nil)
	if err != nil {
		return fmt.Errorf("failed to establish WebSocket connection: %w", err)
	}
	defer conn.Close()

	done := make(chan struct{})

	// Goroutine to read events from WebSocket.
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				// WebSocket error or client disconnected
				fmt.Printf("websocket connection error: %v\n", err)
				return
			}

			var event model.Event
			if err := json.Unmarshal(message, &event); err != nil {
				fmt.Printf("failed to unmarshal WebSocket message: %v\n", err)
				continue
			}

			onEvent(event)
		}
	}()

	// Wait until the connection is closed.
	<-done
	return nil
}
