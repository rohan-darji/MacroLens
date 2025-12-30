package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/macrolens/backend/config"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
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

	handler := NewHandler()
	return SetupRouter(cfg, handler)
}

// TestHealthCheckEndpoint tests the health check endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	router := setupTestRouter()

	t.Run("returns healthy status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "macrolens-backend", response["service"])
		assert.Equal(t, "1.0.0", response["version"])
	})

	t.Run("accepts GET requests only", func(t *testing.T) {
		methods := []string{"POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			req, _ := http.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Method %s should return 404", method)
		}
	})
}

// TestNutritionSearchEndpoint tests the nutrition search endpoint
func TestNutritionSearchEndpoint(t *testing.T) {
	router := setupTestRouter()

	t.Run("returns not implemented status", func(t *testing.T) {
		payload := `{"product_name":"milk","brand":"organic valley"}`
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotImplemented, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response["error"], "not yet implemented")
	})

	t.Run("validates HTTP method", func(t *testing.T) {
		methods := []string{"GET", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			req, _ := http.NewRequest(method, "/api/v1/nutrition/search", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Method %s should return 404", method)
		}
	})

	t.Run("requires correct path", func(t *testing.T) {
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

			assert.Equal(t, http.StatusNotFound, w.Code, "Path %s should return 404", path)
		}
	})
}

// TestCORSIntegration tests CORS headers work end-to-end with full router
func TestCORSIntegration(t *testing.T) {
	router := setupTestRouter()

	t.Run("health endpoint has CORS for Chrome extension", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		req.Header.Set("Origin", "chrome-extension://abcdefghijklmnop")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "chrome-extension://abcdefghijklmnop", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("nutrition endpoint has CORS for localhost", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

// TestRecoveryMiddleware tests panic recovery
func TestRecoveryMiddleware(t *testing.T) {
	router := setupTestRouter()

	// Add a test route that panics
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	t.Run("recovers from panic without crashing server", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()

		// This should not crash the test - recovery middleware should handle it
		router.ServeHTTP(w, req)

		// Gin's default recovery returns 500
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestAPIVersioning tests that API v1 routes are correctly versioned
func TestAPIVersioning(t *testing.T) {
	router := setupTestRouter()

	t.Run("v1 routes are accessible", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/nutrition/search", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 501 Not Implemented, not 404 Not Found
		assert.Equal(t, http.StatusNotImplemented, w.Code)
	})

	t.Run("non-versioned routes return 404", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/nutrition/search", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestJSONResponses tests that all responses are valid JSON
func TestJSONResponses(t *testing.T) {
	router := setupTestRouter()

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"POST", "/api/v1/nutrition/search"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "Response should be valid JSON")
		})
	}
}
