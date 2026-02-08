package usda

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/macrolens/backend/internal/domain"
	"golang.org/x/time/rate"
)

const (
	// maxErrorBodySize limits how much of an error response body we read
	// to prevent memory issues from large error responses
	maxErrorBodySize = 4096
)

// Client handles communication with the USDA FoodData Central API
type Client struct {
	httpClient  *http.Client
	apiKey      string
	baseURL     string
	rateLimiter *rate.Limiter
	debug       bool
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
		debug:       false, // Set to true only for local development
	}
}

// SetDebug enables or disables debug logging
func (c *Client) SetDebug(enabled bool) {
	c.debug = enabled
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
	c.debugLog("SearchFoods called with query: %q", query)

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
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		// Execute request
		resp, err := c.doRequest(ctx, reqURL)
		if err != nil {
			c.debugLog("Request error (attempt %d): %v", attempt, err)
			lastErr = err
			time.Sleep(exponentialBackoff(attempt))
			continue
		}

		// Check status code first before reading body
		if resp.StatusCode != http.StatusOK {
			// Read limited body for error context
			body, readErr := readLimitedBody(resp.Body, maxErrorBodySize)
			resp.Body.Close()

			if readErr != nil {
				c.debugLog("Error reading response body (attempt %d): %v", attempt, readErr)
			}

			c.debugLog("API error (attempt %d) - Status: %d, Body: %s", attempt, resp.StatusCode, string(body))

			if resp.StatusCode == http.StatusNotFound {
				return nil, domain.ErrProductNotFound
			}

			// Retry on server errors (5xx), rate limiting (429), and nginx proxy errors (400 from nginx)
			isNginxProxyError := resp.StatusCode == http.StatusBadRequest && strings.Contains(string(body), "nginx")
			if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests || isNginxProxyError {
				if isNginxProxyError {
					c.debugLog("Nginx proxy error detected (attempt %d), retrying...", attempt)
				}
				lastErr = fmt.Errorf("%w: status %d", domain.ErrUSDAAPIFailure, resp.StatusCode)
				time.Sleep(exponentialBackoff(attempt))
				continue
			}

			// For other 4xx errors, don't retry as it's likely a client error
			return nil, fmt.Errorf("%w: status %d", domain.ErrUSDAAPIFailure, resp.StatusCode)
		}

		// Read successful response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			c.debugLog("Error reading response body: %v", err)
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		// Parse response
		var searchResp domain.USDASearchResponse
		if err := json.Unmarshal(body, &searchResp); err != nil {
			c.debugLog("JSON decode error: %v", err)
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if len(searchResp.Foods) == 0 {
			c.debugLog("No foods found for query: %q", query)
			return nil, domain.ErrProductNotFound
		}

		c.debugLog("Found %d foods for query: %q", len(searchResp.Foods), query)
		return &searchResp, nil
	}

	c.debugLog("All retries failed for query: %q", query)
	return nil, lastErr
}

// debugLog logs a message only when debug mode is enabled
func (c *Client) debugLog(format string, args ...interface{}) {
	if c.debug {
		fmt.Printf("[USDA] "+format+"\n", args...)
	}
}

// exponentialBackoff returns the sleep duration for a given retry attempt
// Uses true exponential backoff: 500ms, 1000ms, 2000ms
func exponentialBackoff(attempt int) time.Duration {
	return time.Duration(500*(1<<(attempt-1))) * time.Millisecond
}

// readLimitedBody reads up to maxBytes from a reader
// This prevents memory issues from large error responses
func readLimitedBody(r io.ReadCloser, maxBytes int64) ([]byte, error) {
	limitedReader := io.LimitReader(r, maxBytes)
	return io.ReadAll(limitedReader)
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
		body, readErr := readLimitedBody(resp.Body, maxErrorBodySize)
		if readErr != nil {
			c.debugLog("Error reading error response body: %v", readErr)
		}
		return nil, fmt.Errorf("%w: status %d, body: %s", domain.ErrUSDAAPIFailure, resp.StatusCode, string(body))
	}

	// Parse response
	var food domain.USDAFood
	if err := json.NewDecoder(resp.Body).Decode(&food); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &food, nil
}
