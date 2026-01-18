package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/macrolens/backend/internal/domain"
	"github.com/macrolens/backend/internal/infrastructure/usda"
)

// Package-level compiled regex patterns for performance
var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9\s]`)
	multipleSpacesRegex  = regexp.MustCompile(`\s+`)
)

// NutritionServiceConfig holds configuration for the nutrition service
type NutritionServiceConfig struct {
	CacheTTL               time.Duration
	MinConfidenceThreshold float64
	EnableFuzzyMatching    bool
	EnableDebugLogging     bool
}

// NutritionService handles nutrition data lookup with caching
type NutritionService struct {
	cache             domain.CacheRepository
	usdaClient        domain.USDAClient
	matchingService   *MatchingService
	queryPreprocessor *QueryPreprocessor
	cacheTTL          time.Duration
}

// NewNutritionService creates a new nutrition service with dependencies
func NewNutritionService(
	cache domain.CacheRepository,
	usdaClient domain.USDAClient,
	config NutritionServiceConfig,
) *NutritionService {
	matchingService := NewMatchingService(MatchConfig{
		MinConfidenceThreshold: config.MinConfidenceThreshold,
		EnableFuzzyMatching:    config.EnableFuzzyMatching,
		EnableDebugLogging:     config.EnableDebugLogging,
	})

	queryPreprocessor := NewQueryPreprocessor(config.EnableDebugLogging)

	cacheTTL := config.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 720 * time.Hour // Default 30 days
	}

	return &NutritionService{
		cache:             cache,
		usdaClient:        usdaClient,
		matchingService:   matchingService,
		queryPreprocessor: queryPreprocessor,
		cacheTTL:          cacheTTL,
	}
}

// SearchNutrition looks up nutrition data for a product.
// Flow: check cache -> search USDA -> match best result -> cache -> return
func (s *NutritionService) SearchNutrition(
	ctx context.Context,
	request *domain.SearchRequest,
) (*domain.NutritionData, error) {
	if request == nil || request.ProductName == "" {
		return nil, domain.ErrInvalidRequest
	}

	cacheKey := s.generateCacheKey(request)

	// Try cache first
	cached, err := s.getFromCache(ctx, cacheKey)
	if err == nil && cached != nil {
		cached.Source = "Cache"
		return cached, nil
	}

	// Cache miss - search USDA with preprocessed query
	query := s.queryPreprocessor.PreprocessQuery(request.ProductName, request.Brand)
	searchResult, err := s.usdaClient.SearchFoods(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrUSDAAPIFailure, err)
	}

	if len(searchResult.Foods) == 0 {
		return nil, domain.ErrProductNotFound
	}

	// Find best match
	matchResult, err := s.matchingService.FindBestMatch(ctx, request, searchResult.Foods)
	if err != nil {
		// For low confidence, still return the data with the error
		if errors.Is(err, domain.ErrLowConfidence) && matchResult != nil {
			nutritionData := s.mapMatchToNutrition(searchResult.Foods, matchResult)
			// Don't cache low confidence results
			return nutritionData, err
		}
		return nil, err
	}

	// Map matched food to NutritionData
	nutritionData := s.mapMatchToNutrition(searchResult.Foods, matchResult)

	// Cache the result
	if err := s.setInCache(ctx, cacheKey, nutritionData); err != nil {
		// Log but don't fail if caching fails
		// In production, this would be logged
	}

	return nutritionData, nil
}

// generateCacheKey creates a normalized cache key from search request.
// Format: "nutrition:{normalized_product_name}:{brand}"
func (s *NutritionService) generateCacheKey(request *domain.SearchRequest) string {
	normalizedName := normalizeForCacheKey(request.ProductName)
	normalizedBrand := normalizeForCacheKey(request.Brand)
	return fmt.Sprintf("nutrition:%s:%s", normalizedName, normalizedBrand)
}

// normalizeForCacheKey normalizes a string for use as cache key component.
// Converts to lowercase, removes special characters, and trims whitespace.
func normalizeForCacheKey(s string) string {
	if s == "" {
		return ""
	}
	result := strings.ToLower(s)
	result = nonAlphanumericRegex.ReplaceAllString(result, "")
	result = multipleSpacesRegex.ReplaceAllString(result, " ")
	return strings.TrimSpace(result)
}

// buildSearchQuery builds a search query string from the request
func buildSearchQuery(request *domain.SearchRequest) string {
	query := request.ProductName
	if request.Brand != "" {
		query = request.Brand + " " + query
	}
	return query
}

// getFromCache retrieves nutrition data from cache
func (s *NutritionService) getFromCache(ctx context.Context, key string) (*domain.NutritionData, error) {
	value, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	nutritionData, ok := value.(*domain.NutritionData)
	if !ok {
		// Try to handle if stored as map
		if dataMap, ok := value.(map[string]interface{}); ok {
			return mapToNutritionData(dataMap), nil
		}
		return nil, domain.ErrCacheMiss
	}

	return nutritionData, nil
}

// setInCache stores nutrition data in cache
func (s *NutritionService) setInCache(ctx context.Context, key string, data *domain.NutritionData) error {
	data.CachedAt = time.Now()
	return s.cache.Set(ctx, key, data, s.cacheTTL)
}

// mapMatchToNutrition finds the matched food and converts it to NutritionData
func (s *NutritionService) mapMatchToNutrition(foods []domain.USDAFood, match *domain.MatchResult) *domain.NutritionData {
	for _, food := range foods {
		if fmt.Sprintf("%d", food.FdcID) == match.FdcID {
			return usda.MapToNutritionData(&food, match.MatchScore)
		}
	}
	// Fallback - shouldn't happen if match came from this food list
	return nil
}

// mapToNutritionData converts a map (from JSON cache) to NutritionData
func mapToNutritionData(data map[string]interface{}) *domain.NutritionData {
	result := &domain.NutritionData{}

	if v, ok := data["fdcId"].(string); ok {
		result.FdcID = v
	}
	if v, ok := data["productName"].(string); ok {
		result.ProductName = v
	}
	if v, ok := data["servingSize"].(string); ok {
		result.ServingSize = v
	}
	if v, ok := data["servingSizeUnit"].(string); ok {
		result.ServingSizeUnit = v
	}
	if v, ok := data["confidence"].(float64); ok {
		result.Confidence = v
	}
	if v, ok := data["source"].(string); ok {
		result.Source = v
	}

	if nutrients, ok := data["nutrients"].(map[string]interface{}); ok {
		if v, ok := nutrients["calories"].(float64); ok {
			result.Nutrients.Calories = v
		}
		if v, ok := nutrients["protein"].(float64); ok {
			result.Nutrients.Protein = v
		}
		if v, ok := nutrients["carbohydrates"].(float64); ok {
			result.Nutrients.Carbohydrates = v
		}
		if v, ok := nutrients["totalFat"].(float64); ok {
			result.Nutrients.TotalFat = v
		}
	}

	return result
}
