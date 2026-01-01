package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/macrolens/backend/internal/domain"
)

// MatchConfig holds configuration for the matching service
type MatchConfig struct {
	MinConfidenceThreshold float64
}

// MatchingService handles fuzzy matching of product names to USDA foods
type MatchingService struct {
	minConfidenceThreshold float64
}

// NewMatchingService creates a new matching service with the given configuration
func NewMatchingService(config MatchConfig) *MatchingService {
	threshold := config.MinConfidenceThreshold
	if threshold <= 0 {
		threshold = 40.0 // Default 40% threshold
	}
	return &MatchingService{
		minConfidenceThreshold: threshold,
	}
}

// FindBestMatch finds the best matching USDA food for a search request.
// Returns the best match with confidence score, or error if no match meets threshold.
func (s *MatchingService) FindBestMatch(
	ctx context.Context,
	request *domain.SearchRequest,
	usdaFoods []domain.USDAFood,
) (*domain.MatchResult, error) {
	if request == nil || request.ProductName == "" {
		return nil, domain.ErrInvalidRequest
	}

	if len(usdaFoods) == 0 {
		return nil, domain.ErrProductNotFound
	}

	var bestMatch *domain.MatchResult
	highestScore := -1.0 // Initialize to -1 so any score (including 0) is considered

	for _, food := range usdaFoods {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		score, matchedTokens := s.calculateMatchScore(request.ProductName, request.Brand, food.Description)

		if score > highestScore {
			highestScore = score
			bestMatch = &domain.MatchResult{
				FdcID:         fmt.Sprintf("%d", food.FdcID),
				Description:   food.Description,
				MatchScore:    score,
				MatchedTokens: matchedTokens,
			}
		}
	}

	if bestMatch == nil {
		return nil, domain.ErrProductNotFound
	}

	if bestMatch.MatchScore < s.minConfidenceThreshold {
		return bestMatch, domain.ErrLowConfidence
	}

	return bestMatch, nil
}

// calculateMatchScore computes similarity between product name and USDA description.
// Uses token-based matching with brand boosting.
// Returns the score (0-100) and the list of matched tokens.
func (s *MatchingService) calculateMatchScore(productName, brand, usdaDescription string) (float64, []string) {
	productTokens := tokenize(productName)
	usdaTokens := tokenize(usdaDescription)

	if len(productTokens) == 0 || len(usdaTokens) == 0 {
		return 0, nil
	}

	// Calculate Jaccard similarity
	intersection, matchedTokens := findIntersection(productTokens, usdaTokens)
	union := findUnion(productTokens, usdaTokens)

	var score float64
	if union > 0 {
		score = (float64(intersection) / float64(union)) * 100
	}

	// Brand matching bonus: +20 points if brand appears in USDA description
	if brand != "" {
		brandLower := strings.ToLower(brand)
		usdaLower := strings.ToLower(usdaDescription)
		if strings.Contains(usdaLower, brandLower) {
			score += 20
		}
	}

	// Exact substring bonus: +15 points for exact product name substring match
	productLower := strings.ToLower(productName)
	usdaLower := strings.ToLower(usdaDescription)
	if strings.Contains(usdaLower, productLower) || strings.Contains(productLower, usdaLower) {
		score += 15
	}

	// Cap score at 100
	if score > 100 {
		score = 100
	}

	return score, matchedTokens
}

// tokenize splits a string into normalized lowercase tokens.
// Removes punctuation and extra whitespace.
func tokenize(s string) []string {
	// Remove punctuation and convert to lowercase
	reg := regexp.MustCompile(`[^\w\s]`)
	cleaned := reg.ReplaceAllString(strings.ToLower(s), " ")

	// Split on whitespace
	words := strings.Fields(cleaned)

	// Filter out common stop words and short tokens
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "or": true,
		"of": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "with": true, "by": true, "from": true,
	}

	var tokens []string
	for _, word := range words {
		if len(word) > 1 && !stopWords[word] {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

// findIntersection returns the count of common tokens and the list of matched tokens
func findIntersection(tokens1, tokens2 []string) (int, []string) {
	set := make(map[string]bool)
	for _, t := range tokens1 {
		set[t] = true
	}

	var matched []string
	seen := make(map[string]bool)
	for _, t := range tokens2 {
		if set[t] && !seen[t] {
			matched = append(matched, t)
			seen[t] = true
		}
	}

	return len(matched), matched
}

// findUnion returns the count of unique tokens across both sets
func findUnion(tokens1, tokens2 []string) int {
	set := make(map[string]bool)
	for _, t := range tokens1 {
		set[t] = true
	}
	for _, t := range tokens2 {
		set[t] = true
	}
	return len(set)
}
