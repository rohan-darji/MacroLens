package usda

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/macrolens/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key", "https://api.example.com")

	assert.NotNil(t, client)
	assert.Equal(t, "test-api-key", client.apiKey)
	assert.Equal(t, "https://api.example.com", client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.rateLimiter)
	assert.False(t, client.debug)
}

func TestSetDebug(t *testing.T) {
	client := NewClient("test-api-key", "https://api.example.com")

	assert.False(t, client.debug)

	client.SetDebug(true)
	assert.True(t, client.debug)

	client.SetDebug(false)
	assert.False(t, client.debug)
}

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 500 * time.Millisecond},
		{2, 1000 * time.Millisecond},
		{3, 2000 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := exponentialBackoff(tt.attempt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchFoods_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/foods/search", r.URL.Path)
		assert.Equal(t, "test-query", r.URL.Query().Get("query"))
		assert.Equal(t, "test-api-key", r.URL.Query().Get("api_key"))

		response := domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{
					FdcID:       123456,
					Description: "Test Food",
					DataType:    "Branded",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "test-query")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Foods, 1)
	assert.Equal(t, 123456, result.Foods[0].FdcID)
	assert.Equal(t, "Test Food", result.Foods[0].Description)
}

func TestSearchFoods_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "nonexistent-product")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrProductNotFound)
}

func TestSearchFoods_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := domain.USDASearchResponse{
			Foods: []domain.USDAFood{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "empty-results")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrProductNotFound)
}

func TestSearchFoods_ServerError_Retries(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{FdcID: 123, Description: "Success after retry"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "retry-test")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, attempts)
}

func TestSearchFoods_ClientError_NoRetry(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "bad-request")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // Should not retry 4xx errors
}

func TestSearchFoods_TooManyRequests_Retries(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		response := domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{FdcID: 456, Description: "Success after rate limit"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "rate-limit-test")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, attempts)
}

func TestSearchFoods_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "invalid-json")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestSearchFoods_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := client.SearchFoods(ctx, "timeout-test")

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestGetFoodDetails_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/food/123456", r.URL.Path)
		assert.Equal(t, "test-api-key", r.URL.Query().Get("api_key"))

		food := domain.USDAFood{
			FdcID:       123456,
			Description: "Detailed Food",
			DataType:    "Branded",
			Nutrients: []domain.USDANutrient{
				{NutrientID: 1003, NutrientName: "Protein", Value: 10.5, UnitName: "G"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(food)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.GetFoodDetails(ctx, "123456")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 123456, result.FdcID)
	assert.Equal(t, "Detailed Food", result.Description)
}

func TestGetFoodDetails_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.GetFoodDetails(ctx, "nonexistent")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrProductNotFound)
}

func TestGetFoodDetails_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.GetFoodDetails(ctx, "error-test")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUSDAAPIFailure)
}

func TestGetFoodDetails_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.GetFoodDetails(ctx, "invalid-json")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestDebugLog(t *testing.T) {
	client := NewClient("test-api-key", "https://api.example.com")

	// Should not panic when debug is false
	client.debug = false
	client.debugLog("test message %s", "arg")

	// Should not panic when debug is true
	client.debug = true
	client.debugLog("test message %s", "arg")
}

func TestReadLimitedBody(t *testing.T) {
	t.Run("reads within limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("short content"))
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := readLimitedBody(resp.Body, 1000)
		require.NoError(t, err)
		assert.Equal(t, "short content", string(body))
	})

	t.Run("truncates beyond limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write more than limit
			for i := 0; i < 100; i++ {
				w.Write([]byte("0123456789"))
			}
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := readLimitedBody(resp.Body, 100)
		require.NoError(t, err)
		assert.Len(t, body, 100)
	})
}

func TestSearchFoods_AllRetriesFail(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test-api-key", server.URL)
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "all-fail")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Equal(t, 3, attempts) // Should try 3 times
}

func TestSearchFoods_RequestCreationError(t *testing.T) {
	client := NewClient("test-api-key", "://invalid-url")
	ctx := context.Background()

	result, err := client.SearchFoods(ctx, "test")

	assert.Nil(t, result)
	assert.Error(t, err)
}

