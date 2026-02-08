package usecase

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/macrolens/backend/internal/domain"
)

// Package-level compiled regex pattern for performance
var punctuationRegex = regexp.MustCompile(`[^\w\s]`)

// Token weight categories for scoring
const (
	weightFood        = 3.0 // Core food terms (milk, chicken, bread)
	weightDescriptive = 2.0 // Descriptive terms (whole, skim, organic)
	weightDefault     = 1.0 // Everything else
	fuzzyWeightFactor = 0.8 // Fuzzy matches get 80% of normal weight
)

// Scoring bonuses
const (
	brandMatchBonus    = 25.0 // Brand appears in USDA description
	substringMatchBonus = 10.0 // Product name is substring of USDA description
	dataTypeBrandedBonus = 10.0 // USDA Branded data type
	dataTypeSurveyBonus  = 5.0  // USDA Survey (FNDDS) data type
	dataTypeFoundationBonus = 3.0 // USDA Foundation data type
	baseScoreMultiplier = 70.0 // Base score max before bonuses
)

// foodTerms contains high-importance food keywords (weight 3.0)
var foodTerms = map[string]bool{
	// Proteins
	"chicken": true, "beef": true, "pork": true, "fish": true, "salmon": true,
	"turkey": true, "lamb": true, "shrimp": true, "tuna": true, "bacon": true,
	"sausage": true, "steak": true, "ham": true, "crab": true, "lobster": true,
	// Dairy
	"milk": true, "cheese": true, "yogurt": true, "butter": true, "cream": true,
	"eggs": true, "egg": true, "cheddar": true, "mozzarella": true, "parmesan": true,
	// Grains
	"bread": true, "rice": true, "pasta": true, "cereal": true, "oats": true,
	"wheat": true, "flour": true, "noodles": true, "tortilla": true, "bagel": true,
	// Produce
	"apple": true, "banana": true, "orange": true, "lettuce": true, "tomato": true,
	"potato": true, "onion": true, "carrot": true, "broccoli": true, "spinach": true,
	"strawberry": true, "blueberry": true, "grape": true, "lemon": true, "lime": true,
	"avocado": true, "cucumber": true, "pepper": true, "corn": true, "beans": true,
	// Beverages
	"juice": true, "soda": true, "cola": true, "coffee": true, "tea": true,
	"water": true, "lemonade": true, "smoothie": true, "shake": true,
	// Snacks & Sweets
	"chips": true, "crackers": true, "cookies": true, "candy": true, "chocolate": true,
	"cake": true, "ice": true, "pie": true, "brownie": true, "popcorn": true,
	// Condiments & Sauces
	"ketchup": true, "mustard": true, "mayo": true, "mayonnaise": true, "sauce": true,
	"salsa": true, "dressing": true, "syrup": true, "honey": true, "jam": true,
	// Prepared Foods
	"pizza": true, "burger": true, "sandwich": true, "soup": true, "salad": true,
	"burrito": true, "taco": true, "wrap": true, "hot": true, "dog": true,
}

// descriptiveTerms contains medium-importance descriptive keywords (weight 2.0)
var descriptiveTerms = map[string]bool{
	// Preparation/processing
	"whole": true, "skim": true, "reduced": true, "fat": true, "low": true,
	"nonfat": true, "organic": true, "natural": true, "fresh": true, "frozen": true,
	"canned": true, "dried": true, "raw": true, "cooked": true, "grilled": true,
	"baked": true, "fried": true, "roasted": true, "smoked": true, "steamed": true,
	// Flavor/variety
	"vanilla": true, "strawberry": true, "plain": true, "flavored": true,
	"original": true, "classic": true, "sweet": true, "spicy": true, "mild": true,
	"hot": true, "regular": true, "lite": true, "light": true, "diet": true,
	// Type descriptors
	"white": true, "brown": true, "refined": true, "enriched": true,
	"fortified": true, "unsweetened": true, "sweetened": true, "salted": true,
	"unsalted": true, "boneless": true, "skinless": true, "lean": true,
	// Nutritional qualifiers
	"vitamin": true, "protein": true, "fiber": true, "calcium": true, "iron": true,
	"omega": true, "probiotic": true, "gluten": true, "free": true, "added": true,
}

// extendedStopWords includes basic English stop words plus product-specific noise
var extendedStopWords = map[string]bool{
	// Basic English stop words
	"a": true, "an": true, "the": true, "and": true, "or": true,
	"of": true, "in": true, "on": true, "at": true, "to": true,
	"for": true, "with": true, "by": true, "from": true, "is": true,
	"it": true, "as": true, "be": true, "was": true, "are": true,
	// Size/quantity units
	"oz": true, "fl": true, "lb": true, "lbs": true, "ml": true,
	"gallon": true, "quart": true, "pint": true, "liter": true, "liters": true,
	"gram": true, "grams": true, "kg": true, "ounce": true, "ounces": true,
	"cup": true, "cups": true, "tbsp": true, "tsp": true,
	// Packaging terms
	"pack": true, "packs": true, "count": true, "ct": true, "pk": true,
	"box": true, "bag": true, "bottle": true, "bottles": true, "can": true,
	"cans": true, "carton": true, "container": true, "pouch": true, "jar": true,
	"tub": true, "sleeve": true, "roll": true, "rolls": true,
	// Marketing/generic terms
	"size": true, "value": true, "family": true, "each": true, "per": true,
	"serving": true, "servings": true, "approx": true, "approximately": true,
	"bonus": true, "new": true, "improved": true, "product": true,
}

// MatchConfig holds configuration for the matching service
type MatchConfig struct {
	MinConfidenceThreshold float64
	EnableFuzzyMatching    bool
	FuzzyEditDistance      int
	EnableDebugLogging     bool
}

// MatchingService handles fuzzy matching of product names to USDA foods
type MatchingService struct {
	minConfidenceThreshold float64
	enableFuzzyMatching    bool
	fuzzyEditDistance      int
	enableDebugLogging     bool
}

// NewMatchingService creates a new matching service with the given configuration
func NewMatchingService(config MatchConfig) *MatchingService {
	threshold := config.MinConfidenceThreshold
	if threshold <= 0 {
		threshold = 40.0 // Default 40% threshold
	}

	fuzzyDist := config.FuzzyEditDistance
	if fuzzyDist <= 0 {
		fuzzyDist = 1 // Default edit distance of 1
	}

	return &MatchingService{
		minConfidenceThreshold: threshold,
		enableFuzzyMatching:    config.EnableFuzzyMatching,
		fuzzyEditDistance:      fuzzyDist,
		enableDebugLogging:     config.EnableDebugLogging,
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

	if s.enableDebugLogging {
		log.Printf("[MATCH] Searching for: %q (brand: %q)", request.ProductName, request.Brand)
	}

	var bestMatch *domain.MatchResult
	highestScore := -1.0 // Initialize to -1 so any score (including 0) is considered

	for _, food := range usdaFoods {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		score, matchedTokens := s.calculateMatchScore(request.ProductName, request.Brand, food.Description, food.DataType)

		if s.enableDebugLogging {
			log.Printf("[MATCH] USDA: %q | DataType: %s | Score: %.1f | Matched: %v",
				food.Description, food.DataType, score, matchedTokens)
		}

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

	if s.enableDebugLogging {
		log.Printf("[MATCH] Best match: %q (confidence: %.1f%%)", bestMatch.Description, bestMatch.MatchScore)
	}

	if bestMatch.MatchScore < s.minConfidenceThreshold {
		return bestMatch, domain.ErrLowConfidence
	}

	return bestMatch, nil
}

// calculateMatchScore computes similarity between product name and USDA description.
// Uses a weighted combination of:
//   - Product token coverage: what % of the product name tokens appear in the USDA result (most important)
//   - USDA token coverage: what % of the USDA description tokens appear in the product name
//   - Brand matching bonus
//   - Substring match bonus
//
// Returns the score (0-100) and the list of matched tokens.
func (s *MatchingService) calculateMatchScore(productName, brand, usdaDescription string) (float64, []string) {
	// Clean the product name first (strip size/noise for better matching)
	cleanedProduct := cleanProductNameForMatching(productName)
	productTokens := tokenize(cleanedProduct)
	usdaTokens := tokenize(usdaDescription)

	if len(productTokens) == 0 || len(usdaTokens) == 0 {
		return 0, nil
	}

	// Product coverage: what fraction of the product tokens are found in USDA description
	// This is the most important signal - if "whole", "milk" both match, that's good
	productMatched, matchedTokens := findIntersection(productTokens, usdaTokens)
	productCoverage := float64(productMatched) / float64(len(productTokens))

	// USDA coverage: what fraction of USDA tokens are found in the product name
	// Lower weight - USDA descriptions contain many extra details
	usdaMatched, _ := findIntersection(usdaTokens, productTokens)
	usdaCoverage := float64(usdaMatched) / float64(len(usdaTokens))

	// Weighted combination: product coverage matters most (60%), USDA coverage (20%), base Jaccard (20%)
	union := findUnion(productTokens, usdaTokens)
	jaccard := float64(productMatched) / float64(union)

	score := (productCoverage*0.60 + usdaCoverage*0.20 + jaccard*0.20) * 100

	// Pre-calculate lowercase versions for bonuses
	productLower := strings.ToLower(cleanedProduct)
	usdaLower := strings.ToLower(usdaDescription)

	// Brand matching bonus: +15 points if brand appears in USDA description
	if brand != "" {
		brandLower := strings.ToLower(brand)
		if strings.Contains(usdaLower, brandLower) {
			score += 15
		}
	}

	// Exact substring bonus: +10 points for exact product name substring match
	if len(productLower) > 3 && (strings.Contains(usdaLower, productLower) || strings.Contains(productLower, usdaLower)) {
		score += 10
	}

	// Cap score at 100
	if score > 100 {
		score = 100
	}

	return score, matchedTokens
}

// cleanProductNameForMatching strips noise from the product name for better token matching.
// This is separate from cleanProductName (used for USDA search query) because
// matching can be more aggressive about stripping noise.
func cleanProductNameForMatching(name string) string {
	// Strip everything after first comma
	if idx := strings.Index(name, ","); idx > 0 {
		name = name[:idx]
	}

	// Remove size patterns
	name = sizePatternRegex.ReplaceAllString(name, " ")

	// Collapse whitespace
	name = multipleSpacesRegex.ReplaceAllString(name, " ")
	return strings.TrimSpace(name)
}

// sizePatternRegex is declared in nutrition_service.go

// tokenize splits a string into normalized lowercase tokens.
// Removes punctuation, stop words, product noise, and pure numeric tokens.
func tokenize(s string) []string {
	// Remove punctuation and convert to lowercase
	cleaned := punctuationRegex.ReplaceAllString(strings.ToLower(s), " ")

	// Split on whitespace
	words := strings.Fields(cleaned)

	var tokens []string
	for _, word := range words {
		// Skip short tokens (1 char or less)
		if len(word) <= 1 {
			continue
		}
		// Skip stop words and product noise
		if extendedStopWords[word] {
			continue
		}
		// Skip pure numeric tokens (e.g., "128", "12")
		if isNumeric(word) {
			continue
		}
		tokens = append(tokens, word)
	}

	return tokens
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// fuzzyTokenMatch checks if two tokens are similar within the edit distance threshold
func fuzzyTokenMatch(token1, token2 string, threshold int) bool {
	// Identical tokens (shouldn't reach here but check anyway)
	if token1 == token2 {
		return true
	}

	// Only apply fuzzy matching to tokens > 4 chars to avoid false positives
	if len(token1) < 4 || len(token2) < 4 {
		return false
	}

	// Quick length check - if lengths differ by more than threshold, can't match
	lenDiff := len(token1) - len(token2)
	if lenDiff < 0 {
		lenDiff = -lenDiff
	}
	if lenDiff > threshold {
		return false
	}

	return levenshteinDistance(token1, token2) <= threshold
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	r1 := []rune(s1)
	r2 := []rune(s2)
	m := len(r1)
	n := len(r2)

	// Use two rows instead of full matrix for space efficiency
	prev := make([]int, n+1)
	curr := make([]int, n+1)

	// Initialize first row
	for j := 0; j <= n; j++ {
		prev[j] = j
	}

	// Fill matrix
	for i := 1; i <= m; i++ {
		curr[0] = i
		for j := 1; j <= n; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			curr[j] = min(
				prev[j]+1,      // deletion
				curr[j-1]+1,    // insertion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[n]
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
