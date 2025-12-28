package domain

import "errors"

var (
	// ErrProductNotFound is returned when a product cannot be found in USDA database
	ErrProductNotFound = errors.New("product not found in USDA database")

	// ErrLowConfidence is returned when the match confidence is below the threshold
	ErrLowConfidence = errors.New("match confidence below threshold")

	// ErrRateLimited is returned when rate limit is exceeded
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrInvalidRequest is returned when request parameters are invalid
	ErrInvalidRequest = errors.New("invalid request parameters")

	// ErrCacheMiss is returned when data is not found in cache
	ErrCacheMiss = errors.New("cache miss")

	// ErrUSDAAPIFailure is returned when USDA API request fails
	ErrUSDAAPIFailure = errors.New("USDA API request failed")

	// ErrCacheUnavailable is returned when cache service is unavailable
	ErrCacheUnavailable = errors.New("cache service unavailable")
)
