package usda

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/macrolens/backend/internal/domain"
	"golang.org/x/time/rate"
)

// Client handles communication with the USDA FoodData Central API
type Client struct {
	httpClient  *http.Client
	apiKey      string
	baseURL     string
	rateLimiter *rate.Limiter
}

// NewClient creates a new USDA API client
func NewClient(apiKey, baseURL string) *Client {
	// USDA allows 1000 requests per hour
	// rate.Limit is requests per second, so 1000/3600 ≈ 0.278 requests/sec
	limiter := rate.NewLimiter(rate.Limit(0.278), 10) // burst of 10 requests

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:      apiKey,
		baseURL:     baseURL,
		rateLimiter: limiter,
	}
}

// SearchFoods searches for foods in the USDA database
func (c *Client) SearchFoods(ctx context.Context, query string) (*domain.USDASearchResponse, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build request URL
	endpoint := fmt.Sprintf("%s/v1/foods/search", c.baseURL)
	params := url.Values{}
	params.Add("query", query)
	params.Add("api_key", c.apiKey)
	params.Add("dataType", "Survey (FNDDS),Foundation,Branded") // Focus on relevant data types
	params.Add("pageSize", "10") // Get top 10 results

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrUSDAAPIFailure, err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", domain.ErrUSDAAPIFailure, resp.StatusCode, string(body))
	}

	// Parse response
	var searchResp domain.USDASearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.Foods) == 0 {
		return nil, domain.ErrProductNotFound
	}

	return &searchResp, nil
}

// GetFoodDetails retrieves detailed nutrition information for a specific food by FDC ID
func (c *Client) GetFoodDetails(ctx context.Context, fdcID string) (*domain.USDAFood, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build request URL
	endpoint := fmt.Sprintf("%s/v1/food/%s", c.baseURL, fdcID)
	params := url.Values{}
	params.Add("api_key", c.apiKey)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrUSDAAPIFailure, err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrProductNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", domain.ErrUSDAAPIFailure, resp.StatusCode, string(body))
	}

	// Parse response
	var food domain.USDAFood
	if err := json.NewDecoder(resp.Body).Decode(&food); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &food, nil
}
