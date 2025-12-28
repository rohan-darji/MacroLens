package domain

import (
	"context"
	"time"
)

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// USDAClient defines the interface for interacting with USDA FoodData Central API
type USDAClient interface {
	SearchFoods(ctx context.Context, query string) (*USDASearchResponse, error)
	GetFoodDetails(ctx context.Context, fdcID string) (*USDAFood, error)
}

// NutritionRepository defines the interface for nutrition data persistence
// (Future use: could be used for custom nutrition database)
type NutritionRepository interface {
	GetByFdcID(ctx context.Context, fdcID string) (*NutritionData, error)
	Save(ctx context.Context, data *NutritionData) error
}
