package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Message    string
	Response   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

func NewAPIClient(baseURL string, timeout time.Duration) *APIClient {
	slog.Info("Creating new API client", "base_url", baseURL, "timeout", timeout)
	
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *APIClient) Get(ctx context.Context, path string) ([]byte, error) {
	return c.makeRequest(ctx, "GET", path, nil)
}

func (c *APIClient) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.makeRequest(ctx, "POST", path, body)
}

func (c *APIClient) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.makeRequest(ctx, "PUT", path, body)
}

func (c *APIClient) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.makeRequest(ctx, "DELETE", path, nil)
}

func (c *APIClient) makeRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := c.baseURL + path
	
	slog.Info("Making API request", "method", method, "url", url)
	
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			slog.Error("Failed to marshal request body", "error", err)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
		slog.Debug("Request body", "body", string(jsonBody))
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		slog.Error("Failed to create HTTP request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("HTTP request failed", "error", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	slog.Info("API request completed", 
		"status_code", resp.StatusCode,
		"response_size", len(respBody),
	)
	
	if resp.StatusCode >= 400 {
		slog.Error("API request failed", 
			"status_code", resp.StatusCode,
			"response", string(respBody),
		)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			Response:   string(respBody),
		}
	}
	
	slog.Debug("Response body", "body", string(respBody))
	return respBody, nil
}