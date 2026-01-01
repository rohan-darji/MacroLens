package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/macrolens/backend/config"
	"github.com/macrolens/backend/internal/domain"
	"github.com/macrolens/backend/internal/usecase"
)

// TestMain sets up test environment before running tests
func TestMain(m *testing.M) {
	// Set Gin to test mode once for all tests
	gin.SetMode(gin.TestMode)

	// Run tests
	exitCode := m.Run()

	// Exit with the test result code
	os.Exit(exitCode)
}

// setupTestRouter creates a test router with default configuration
func setupTestRouter() *gin.Engine {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:           "8080",
			Environment:    "test",
			AllowedOrigins: []string{"chrome-extension://*", "http://localhost:3000"},
		},
		USDA: config.USDAConfig{
			APIKey:  "test-api-key",
			BaseURL: "https://api.nal.usda.gov/fdc",
		},
		Cache: config.CacheConfig{
			Type: "memory",
		},
	}

	// Pass nil for nutrition service - handler returns 501 for nutrition endpoints
	handler := NewHandler(nil)
	if handler == nil {
		panic("setupTestRouter: NewHandler returned nil")
	}

	router := SetupRouter(cfg, handler)
	if router == nil {
		panic("setupTestRouter: SetupRouter returned nil *gin.Engine")
	}

	return router
}

// TestHealthCheckEndpoint tests the health check endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	t.Run("returns healthy status", func(t *testing.T) {
		router := setupTestRouter()

		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["status"] != "healthy" {
			t.Errorf("status = %v, want healthy", response["status"])
		}
		if response["service"] != "macrolens-backend" {
			t.Errorf("service = %v, want macrolens-backend", response["service"])
		}
		version, ok := response["version"].(string)
		if !ok || strings.TrimSpace(version) == "" {
			t.Errorf("version = %v, want non-empty string", response["version"])
		}
	})

	t.Run("accepts GET requests only", func(t *testing.T) {
		router := setupTestRouter()

		methods := []string{"POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			req, _ := http.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Method %s: Status = %d, want %d", method, w.Code, http.StatusNotFound)
			}
		}
	})
}

// TestNutritionSearchEndpoint tests the nutrition search endpoint
func TestNutritionSearchEndpoint(t *testing.T) {
	t.Run("returns not implemented status", func(t *testing.T) {
		router := setupTestRouter()

		payload := `{"product_name":"milk","brand":"organic valley"}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotImplemented {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusNotImplemented)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		errorMsg, ok := response["error"].(string)
		if !ok {
			t.Errorf("error field is not a string: %v", response["error"])
		} else if !strings.Contains(errorMsg, "not configured") {
			t.Errorf("error = %q, want to contain 'not configured'", errorMsg)
		}
	})

	t.Run("validates HTTP method", func(t *testing.T) {
		router := setupTestRouter()

		methods := []string{"GET", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			req, _ := http.NewRequest(method, "/api/v1/nutrition/search", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Method %s: Status = %d, want %d", method, w.Code, http.StatusNotFound)
			}
		}
	})

	t.Run("requires correct path", func(t *testing.T) {
		router := setupTestRouter()

		incorrectPaths := []string{
			"/api/v1/nutrition",
			"/api/v1/nutrition/",
			"/api/nutrition/search",
			"/nutrition/search",
		}

		for _, path := range incorrectPaths {
			req, _ := http.NewRequest("POST", path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Path %s: Status = %d, want %d", path, w.Code, http.StatusNotFound)
			}
		}
	})
}

// TestCORSIntegration tests CORS headers work end-to-end with full router
func TestCORSIntegration(t *testing.T) {
	t.Run("health endpoint has CORS for Chrome extension", func(t *testing.T) {
		router := setupTestRouter()

		req, _ := http.NewRequest("GET", "/health", nil)
		req.Header.Set("Origin", "chrome-extension://abcdefghijklmnop")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		}

		gotOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if gotOrigin != "chrome-extension://abcdefghijklmnop" {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", gotOrigin, "chrome-extension://abcdefghijklmnop")
		}

		gotCreds := w.Header().Get("Access-Control-Allow-Credentials")
		if gotCreds != "true" {
			t.Errorf("Access-Control-Allow-Credentials = %q, want %q", gotCreds, "true")
		}
	})

	t.Run("nutrition endpoint has CORS for localhost", func(t *testing.T) {
		router := setupTestRouter()

		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		gotOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if gotOrigin != "http://localhost:3000" {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", gotOrigin, "http://localhost:3000")
		}
	})
}

// TestRecoveryMiddleware tests panic recovery
func TestRecoveryMiddleware(t *testing.T) {
	t.Run("recovers from panic without crashing server", func(t *testing.T) {
		router := setupTestRouter()

		// Add a test route that panics
		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req, _ := http.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()

		// This should not crash the test - recovery middleware should handle it
		router.ServeHTTP(w, req)

		// Gin's default recovery returns 500
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

// TestAPIVersioning tests that API v1 routes are correctly versioned
func TestAPIVersioning(t *testing.T) {
	t.Run("v1 routes are accessible", func(t *testing.T) {
		router := setupTestRouter()

		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 501 Not Implemented, not 404 Not Found
		if w.Code != http.StatusNotImplemented {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusNotImplemented)
		}
	})

	t.Run("non-versioned routes return 404", func(t *testing.T) {
		router := setupTestRouter()

		req, _ := http.NewRequest("POST", "/api/nutrition/search", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

// TestJSONResponses tests that all responses are valid JSON
func TestJSONResponses(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"POST", "/api/v1/nutrition/search"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			router := setupTestRouter()

			req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			gotContentType := w.Header().Get("Content-Type")
			wantContentType := "application/json; charset=utf-8"
			if gotContentType != wantContentType {
				t.Errorf("Content-Type = %q, want %q", gotContentType, wantContentType)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Response should be valid JSON, got error: %v", err)
			}
		})
	}
}

// --- Mock implementations for testing with NutritionService ---

// mockCacheRepository is a mock implementation of domain.CacheRepository
type mockCacheRepository struct {
	data map[string]interface{}
}

func newMockCacheRepository() *mockCacheRepository {
	return &mockCacheRepository{data: make(map[string]interface{})}
}

func (m *mockCacheRepository) Get(ctx context.Context, key string) (interface{}, error) {
	if value, ok := m.data[key]; ok {
		return value, nil
	}
	return nil, domain.ErrCacheMiss
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

// mockUSDAClient is a mock implementation of domain.USDAClient
type mockUSDAClient struct {
	searchResult *domain.USDASearchResponse
	searchError  error
}

func newMockUSDAClient() *mockUSDAClient {
	return &mockUSDAClient{}
}

func (m *mockUSDAClient) SearchFoods(ctx context.Context, query string) (*domain.USDASearchResponse, error) {
	if m.searchError != nil {
		return nil, m.searchError
	}
	return m.searchResult, nil
}

func (m *mockUSDAClient) GetFoodDetails(ctx context.Context, fdcID string) (*domain.USDAFood, error) {
	return nil, nil
}

// setupTestRouterWithService creates a test router with a real NutritionService using mocks
func setupTestRouterWithService(cache domain.CacheRepository, client domain.USDAClient) *gin.Engine {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:           "8080",
			Environment:    "test",
			AllowedOrigins: []string{"chrome-extension://*", "http://localhost:3000"},
		},
	}

	nutritionService := usecase.NewNutritionService(
		cache,
		client,
		usecase.NutritionServiceConfig{
			CacheTTL:               24 * time.Hour,
			MinConfidenceThreshold: 40,
		},
	)

	handler := NewHandler(nutritionService)
	return SetupRouter(cfg, handler)
}

// TestNutritionSearchWithService tests the nutrition search endpoint with a real service
func TestNutritionSearchWithService(t *testing.T) {
	t.Run("returns nutrition data for valid request", func(t *testing.T) {
		cache := newMockCacheRepository()
		client := newMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{
					FdcID:       12345,
					Description: "Whole Milk",
					Nutrients: []domain.USDANutrient{
						{NutrientID: 1008, Value: 150}, // Calories
						{NutrientID: 1003, Value: 8},   // Protein
						{NutrientID: 1005, Value: 12},  // Carbs
						{NutrientID: 1004, Value: 8},   // Fat
					},
				},
			},
		}

		router := setupTestRouterWithService(cache, client)

		payload := `{"productName":"whole milk"}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["fdcId"] != "12345" {
			t.Errorf("fdcId = %v, want 12345", response["fdcId"])
		}
		if response["source"] != "USDA" {
			t.Errorf("source = %v, want USDA", response["source"])
		}
	})

	t.Run("returns 400 for missing productName", func(t *testing.T) {
		cache := newMockCacheRepository()
		client := newMockUSDAClient()

		router := setupTestRouterWithService(cache, client)

		payload := `{"brand":"Great Value"}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] == nil {
			t.Error("expected error field in response")
		}
	})

	t.Run("returns 404 when no products found", func(t *testing.T) {
		cache := newMockCacheRepository()
		client := newMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{}, // Empty results
		}

		router := setupTestRouterWithService(cache, client)

		payload := `{"productName":"nonexistent product xyz123"}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		cache := newMockCacheRepository()
		client := newMockUSDAClient()

		router := setupTestRouterWithService(cache, client)

		payload := `{invalid json}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("includes brand in search and response", func(t *testing.T) {
		cache := newMockCacheRepository()
		client := newMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{
					FdcID:       67890,
					Description: "Great Value Whole Milk",
					Nutrients:   []domain.USDANutrient{},
				},
			},
		}

		router := setupTestRouterWithService(cache, client)

		payload := `{"productName":"whole milk","brand":"Great Value"}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("returns 502 for USDA API failure", func(t *testing.T) {
		cache := newMockCacheRepository()
		client := newMockUSDAClient()
		client.searchError = domain.ErrUSDAAPIFailure

		router := setupTestRouterWithService(cache, client)

		payload := `{"productName":"whole milk"}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadGateway {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusBadGateway)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "USDA API temporarily unavailable" {
			t.Errorf("error = %v, want 'USDA API temporarily unavailable'", response["error"])
		}
	})

	t.Run("returns low confidence warning with data", func(t *testing.T) {
		cache := newMockCacheRepository()
		client := newMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{
					FdcID:       99999,
					Description: "Some Unrelated Food",
					Nutrients: []domain.USDANutrient{
						{NutrientID: 1008, Value: 100},
					},
				},
			},
		}

		router := setupTestRouterWithService(cache, client)

		// Request for chocolate cake but USDA returns chicken - low confidence match
		payload := `{"productName":"chocolate cake deluxe premium"}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Low confidence still returns 200 but with warning
		if w.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["warning"] == nil {
			t.Error("expected warning field in response for low confidence match")
		}

		if response["data"] == nil {
			t.Error("expected data field even with low confidence")
		}

		warningStr, ok := response["warning"].(string)
		if !ok || warningStr != "Low confidence match - verify the product manually" {
			t.Errorf("warning = %v, want 'Low confidence match - verify the product manually'", response["warning"])
		}
	})
}
