package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/macrolens/backend/internal/domain"
)

// MockCacheRepository is a mock implementation of domain.CacheRepository
type MockCacheRepository struct {
	data      map[string]interface{}
	getError  error
	setError  error
	getCalled bool
	setCalled bool
}

func NewMockCacheRepository() *MockCacheRepository {
	return &MockCacheRepository{
		data: make(map[string]interface{}),
	}
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (interface{}, error) {
	m.getCalled = true
	if m.getError != nil {
		return nil, m.getError
	}
	if value, ok := m.data[key]; ok {
		return value, nil
	}
	return nil, domain.ErrCacheMiss
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.setCalled = true
	if m.setError != nil {
		return m.setError
	}
	m.data[key] = value
	return nil
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

// MockUSDAClient is a mock implementation of domain.USDAClient
type MockUSDAClient struct {
	searchResult *domain.USDASearchResponse
	searchError  error
	foodResult   *domain.USDAFood
	foodError    error
}

func NewMockUSDAClient() *MockUSDAClient {
	return &MockUSDAClient{}
}

func (m *MockUSDAClient) SearchFoods(ctx context.Context, query string) (*domain.USDASearchResponse, error) {
	if m.searchError != nil {
		return nil, m.searchError
	}
	return m.searchResult, nil
}

func (m *MockUSDAClient) GetFoodDetails(ctx context.Context, fdcID string) (*domain.USDAFood, error) {
	if m.foodError != nil {
		return nil, m.foodError
	}
	return m.foodResult, nil
}

func TestNewNutritionService(t *testing.T) {
	cache := NewMockCacheRepository()
	client := NewMockUSDAClient()

	t.Run("creates service with default values", func(t *testing.T) {
		svc := NewNutritionService(cache, client, NutritionServiceConfig{})
		if svc == nil {
			t.Fatal("expected service to be created")
		}
		if svc.cacheTTL != 720*time.Hour {
			t.Errorf("cacheTTL = %v, want 720h", svc.cacheTTL)
		}
	})

	t.Run("creates service with custom values", func(t *testing.T) {
		svc := NewNutritionService(cache, client, NutritionServiceConfig{
			CacheTTL:               24 * time.Hour,
			MinConfidenceThreshold: 50,
		})
		if svc.cacheTTL != 24*time.Hour {
			t.Errorf("cacheTTL = %v, want 24h", svc.cacheTTL)
		}
	})
}

func TestSearchNutrition(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error for nil request", func(t *testing.T) {
		cache := NewMockCacheRepository()
		client := NewMockUSDAClient()
		svc := NewNutritionService(cache, client, NutritionServiceConfig{})

		_, err := svc.SearchNutrition(ctx, nil)
		if !errors.Is(err, domain.ErrInvalidRequest) {
			t.Errorf("error = %v, want ErrInvalidRequest", err)
		}
	})

	t.Run("returns error for empty product name", func(t *testing.T) {
		cache := NewMockCacheRepository()
		client := NewMockUSDAClient()
		svc := NewNutritionService(cache, client, NutritionServiceConfig{})

		_, err := svc.SearchNutrition(ctx, &domain.SearchRequest{ProductName: ""})
		if !errors.Is(err, domain.ErrInvalidRequest) {
			t.Errorf("error = %v, want ErrInvalidRequest", err)
		}
	})

	t.Run("returns cached data on cache hit", func(t *testing.T) {
		cache := NewMockCacheRepository()
		cachedData := &domain.NutritionData{
			FdcID:       "123",
			ProductName: "Whole Milk",
			Nutrients: domain.Nutrients{
				Calories: 150,
				Protein:  8,
			},
			Confidence: 85,
			Source:     "USDA",
		}
		cache.data["nutrition:whole milk:"] = cachedData

		client := NewMockUSDAClient()
		svc := NewNutritionService(cache, client, NutritionServiceConfig{})

		result, err := svc.SearchNutrition(ctx, &domain.SearchRequest{ProductName: "whole milk"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Source != "Cache" {
			t.Errorf("Source = %v, want Cache", result.Source)
		}
		if result.FdcID != "123" {
			t.Errorf("FdcID = %v, want 123", result.FdcID)
		}
	})

	t.Run("searches USDA on cache miss", func(t *testing.T) {
		cache := NewMockCacheRepository()
		cache.getError = domain.ErrCacheMiss

		client := NewMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{
					FdcID:       456,
					Description: "Whole Milk",
					Nutrients: []domain.USDANutrient{
						{NutrientID: 1008, Value: 150}, // Calories
						{NutrientID: 1003, Value: 8},   // Protein
					},
				},
			},
		}

		svc := NewNutritionService(cache, client, NutritionServiceConfig{
			MinConfidenceThreshold: 40,
		})

		result, err := svc.SearchNutrition(ctx, &domain.SearchRequest{ProductName: "whole milk"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Source != "USDA" {
			t.Errorf("Source = %v, want USDA", result.Source)
		}
		if result.FdcID != "456" {
			t.Errorf("FdcID = %v, want 456", result.FdcID)
		}
		if !cache.setCalled {
			t.Error("expected cache.Set to be called")
		}
	})

	t.Run("returns error when USDA API fails", func(t *testing.T) {
		cache := NewMockCacheRepository()
		cache.getError = domain.ErrCacheMiss

		client := NewMockUSDAClient()
		client.searchError = errors.New("API timeout")

		svc := NewNutritionService(cache, client, NutritionServiceConfig{})

		_, err := svc.SearchNutrition(ctx, &domain.SearchRequest{ProductName: "whole milk"})
		if !errors.Is(err, domain.ErrUSDAAPIFailure) {
			t.Errorf("error = %v, want ErrUSDAAPIFailure", err)
		}
	})

	t.Run("returns error when no products found", func(t *testing.T) {
		cache := NewMockCacheRepository()
		cache.getError = domain.ErrCacheMiss

		client := NewMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{},
		}

		svc := NewNutritionService(cache, client, NutritionServiceConfig{})

		_, err := svc.SearchNutrition(ctx, &domain.SearchRequest{ProductName: "nonexistent product xyz"})
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Errorf("error = %v, want ErrProductNotFound", err)
		}
	})

	t.Run("returns low confidence error with data for poor match", func(t *testing.T) {
		cache := NewMockCacheRepository()
		cache.getError = domain.ErrCacheMiss

		client := NewMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{
					FdcID:       789,
					Description: "Grilled Chicken Breast",
					Nutrients: []domain.USDANutrient{
						{NutrientID: 1008, Value: 165},
					},
				},
			},
		}

		svc := NewNutritionService(cache, client, NutritionServiceConfig{
			MinConfidenceThreshold: 80, // High threshold
		})

		result, err := svc.SearchNutrition(ctx, &domain.SearchRequest{ProductName: "chocolate cake"})
		if !errors.Is(err, domain.ErrLowConfidence) {
			t.Errorf("error = %v, want ErrLowConfidence", err)
		}
		if result == nil {
			t.Error("expected result to be returned even with low confidence")
		}
		if cache.setCalled {
			t.Error("low confidence results should not be cached")
		}
	})

	t.Run("includes brand in search query", func(t *testing.T) {
		cache := NewMockCacheRepository()
		cache.getError = domain.ErrCacheMiss

		client := NewMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{
					FdcID:       111,
					Description: "Great Value Whole Milk",
					Nutrients: []domain.USDANutrient{
						{NutrientID: 1008, Value: 150},
					},
				},
			},
		}

		svc := NewNutritionService(cache, client, NutritionServiceConfig{})

		result, err := svc.SearchNutrition(ctx, &domain.SearchRequest{
			ProductName: "whole milk",
			Brand:       "Great Value",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.FdcID != "111" {
			t.Errorf("FdcID = %v, want 111", result.FdcID)
		}
	})

	t.Run("continues even if caching fails", func(t *testing.T) {
		cache := NewMockCacheRepository()
		cache.getError = domain.ErrCacheMiss
		cache.setError = errors.New("cache write failed")

		client := NewMockUSDAClient()
		client.searchResult = &domain.USDASearchResponse{
			Foods: []domain.USDAFood{
				{
					FdcID:       222,
					Description: "Whole Milk",
					Nutrients:   []domain.USDANutrient{},
				},
			},
		}

		svc := NewNutritionService(cache, client, NutritionServiceConfig{})

		result, err := svc.SearchNutrition(ctx, &domain.SearchRequest{ProductName: "whole milk"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result even when cache write fails")
		}
	})
}

func TestGenerateCacheKey(t *testing.T) {
	cache := NewMockCacheRepository()
	client := NewMockUSDAClient()
	svc := NewNutritionService(cache, client, NutritionServiceConfig{})

	t.Run("generates key with product name only", func(t *testing.T) {
		key := svc.generateCacheKey(&domain.SearchRequest{ProductName: "Whole Milk"})
		if key != "nutrition:whole milk:" {
			t.Errorf("key = %v, want nutrition:whole milk:", key)
		}
	})

	t.Run("generates key with product name and brand", func(t *testing.T) {
		key := svc.generateCacheKey(&domain.SearchRequest{
			ProductName: "Whole Milk",
			Brand:       "Great Value",
		})
		if key != "nutrition:whole milk:great value" {
			t.Errorf("key = %v, want nutrition:whole milk:great value", key)
		}
	})

	t.Run("normalizes special characters", func(t *testing.T) {
		key := svc.generateCacheKey(&domain.SearchRequest{
			ProductName: "2% Milk (Vitamin D)",
			Brand:       "Store-Brand!",
		})
		// Should remove special chars and normalize
		if key != "nutrition:2 milk vitamin d:storebrand" {
			t.Errorf("key = %v, want nutrition:2 milk vitamin d:storebrand", key)
		}
	})
}

func TestNormalizeForCacheKey(t *testing.T) {
	t.Run("converts to lowercase", func(t *testing.T) {
		result := normalizeForCacheKey("WHOLE MILK")
		if result != "whole milk" {
			t.Errorf("result = %v, want 'whole milk'", result)
		}
	})

	t.Run("removes special characters", func(t *testing.T) {
		result := normalizeForCacheKey("milk, 2% (reduced fat)")
		if result != "milk 2 reduced fat" {
			t.Errorf("result = %v, want 'milk 2 reduced fat'", result)
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := normalizeForCacheKey("")
		if result != "" {
			t.Errorf("result = %v, want empty string", result)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		result := normalizeForCacheKey("  milk  ")
		if result != "milk" {
			t.Errorf("result = %v, want 'milk'", result)
		}
	})

	t.Run("collapses multiple spaces", func(t *testing.T) {
		result := normalizeForCacheKey("whole    milk")
		if result != "whole milk" {
			t.Errorf("result = %v, want 'whole milk'", result)
		}
	})
}

func TestBuildSearchQuery(t *testing.T) {
	t.Run("uses product name only when no brand", func(t *testing.T) {
		query := buildSearchQuery(&domain.SearchRequest{ProductName: "whole milk"})
		if query != "whole milk" {
			t.Errorf("query = %v, want 'whole milk'", query)
		}
	})

	t.Run("prepends brand to product name", func(t *testing.T) {
		query := buildSearchQuery(&domain.SearchRequest{
			ProductName: "whole milk",
			Brand:       "Great Value",
		})
		if query != "Great Value whole milk" {
			t.Errorf("query = %v, want 'Great Value whole milk'", query)
		}
	})
}
