package usda

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// doRequest executes an HTTP GET request with proper headers and error handling
func (c *Client) doRequest(ctx context.Context, reqURL string) (*http.Response, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "MacroLens/1.0")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrUSDAAPIFailure, err)
	}

	return resp, nil
}

// SearchFoods searches for foods in the USDA database
func (c *Client) SearchFoods(ctx context.Context, query string) (*domain.USDASearchResponse, error) {
	log.Printf("[USDA] SearchFoods called with query: %q", query)

	// Build request URL
	endpoint := fmt.Sprintf("%s/v1/foods/search", c.baseURL)
	params := url.Values{}
	params.Add("query", query)
	params.Add("api_key", c.apiKey)
	params.Add("dataType", "Survey (FNDDS),Foundation,Branded") // Focus on relevant data types
	params.Add("pageSize", "10")                                // Get top 10 results

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Retry up to 3 times for transient failures
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		// Wait for rate limiter
		if err := c.rateLimiter.Wait(ctx); err != nil {
			log.Printf("[USDA] Rate limiter error: %v", err)
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		// Execute request
		resp, err := c.doRequest(ctx, reqURL)
		if err != nil {
			log.Printf("[USDA] Request error (attempt %d): %v", attempt, err)
			lastErr = err
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
			continue
		}

		// Read body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check status code - retry on 4xx/5xx errors (except 404)
		if resp.StatusCode != http.StatusOK {
			log.Printf("[USDA] API error (attempt %d) - Status: %d, Body: %s", attempt, resp.StatusCode, string(body))
			if resp.StatusCode == http.StatusNotFound {
				return nil, domain.ErrProductNotFound
			}
			lastErr = fmt.Errorf("%w: status %d", domain.ErrUSDAAPIFailure, resp.StatusCode)
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
			continue
		}

		// Parse response
		var searchResp domain.USDASearchResponse
		if err := json.Unmarshal(body, &searchResp); err != nil {
			log.Printf("[USDA] JSON decode error: %v", err)
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if len(searchResp.Foods) == 0 {
			log.Printf("[USDA] No foods found for query: %q", query)
			return nil, domain.ErrProductNotFound
		}

		log.Printf("[USDA] Found %d foods for query: %q", len(searchResp.Foods), query)
		return &searchResp, nil
	}

	log.Printf("[USDA] All retries failed for query: %q", query)
	return nil, lastErr
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

	// Execute request
	resp, err := c.doRequest(ctx, reqURL)
	if err != nil {
		return nil, err
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
