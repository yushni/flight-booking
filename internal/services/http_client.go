package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"flight-booking/internal/interfaces"
)

// HTTPClientService implements the HTTPClient interface
type HTTPClientService struct {
	client *http.Client
}

// NewHTTPClientService creates a new HTTP client service
func NewHTTPClientService(timeout time.Duration) interfaces.HTTPClient {
	return &HTTPClientService{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Get performs a GET request
func (h *HTTPClientService) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return h.doRequest(ctx, http.MethodGet, url, nil, headers)
}

// Post performs a POST request
func (h *HTTPClientService) Post(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error) {
	return h.doRequest(ctx, http.MethodPost, url, body, headers)
}

// Put performs a PUT request
func (h *HTTPClientService) Put(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error) {
	return h.doRequest(ctx, http.MethodPut, url, body, headers)
}

// Delete performs a DELETE request
func (h *HTTPClientService) Delete(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return h.doRequest(ctx, http.MethodDelete, url, nil, headers)
}

// doRequest performs the actual HTTP request
func (h *HTTPClientService) doRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", "flight-booking/1.0")
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}
