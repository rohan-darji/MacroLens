package usecase

import (
	"log"
	"regexp"
	"strings"
)

// QueryPreprocessor handles cleaning and extracting keywords from product names
type QueryPreprocessor struct {
	enableDebugLogging bool
}

// Compiled regex patterns for query preprocessing
var (
	// Matches size/quantity patterns like "128 fl oz", "12 oz", "1.5 liter", "2 lb"
	sizeQuantityPattern = regexp.MustCompile(`\b\d+\.?\d*\s*(fl\s*)?oz\b|\b\d+\.?\d*\s*(fl\s*)?ounces?\b|\b\d+\.?\d*\s*lbs?\b|\b\d+\.?\d*\s*pounds?\b|\b\d+\.?\d*\s*ml\b|\b\d+\.?\d*\s*liters?\b|\b\d+\.?\d*\s*gallons?\b|\b\d+\.?\d*\s*quarts?\b|\b\d+\.?\d*\s*pints?\b|\b\d+\.?\d*\s*kg\b|\b\d+\.?\d*\s*grams?\b|\b\d+\.?\d*\s*g\b`)

	// Matches pack/count patterns like "12 pack", "pack of 6", "6-pack", "24 count", "6 ct", "12 pack cans"
	packCountPattern = regexp.MustCompile(`\b\d+[-\s]*(pack|pk|count|ct)(\s+\w+)?\b|\bpack\s*of\s*\d+\b|\b\d+\s*cans?\b|\b\d+\s*bottles?\b|\b\d+\s*pouches?\b|\b\d+\s*bars?\b|\b\d+\s*pieces?\b`)

	// Matches standalone numbers with no unit (e.g., ", 128", "- 12")
	standaloneNumberPattern = regexp.MustCompile(`[,\-]\s*\d+\.?\d*\s*$|^\d+\.?\d*\s*[,\-]`)

	// Multiple spaces cleanup
	multiSpacePattern = regexp.MustCompile(`\s+`)
)

// noiseWords to remove from queries (marketing terms, generic descriptors)
var queryNoiseWords = map[string]bool{
	// Marketing terms
	"value":     true,
	"family":    true,
	"bonus":     true,
	"new":       true,
	"improved":  true,
	"premium":   true,
	"select":    true,
	"choice":    true,
	"quality":   true,
	"best":      true,
	"great":     true,
	"delicious": true,
	"tasty":     true,
	"favorite":  true,
	"special":   true,

	// Size descriptors
	"size":   true,
	"large":  true,
	"medium": true,
	"small":  true,
	"mini":   true,
	"jumbo":  true,
	"giant":  true,
	"big":    true,
	"snack":  true,
	"single": true,
	"double": true,
	"triple": true,

	// Packaging terms
	"package": true,
	"box":     true,
	"bag":     true,
	"bottle":  true,
	"can":     true,
	"jar":     true,
	"tub":     true,
	"carton":  true,
	"sleeve":  true,
	"pouch":   true,
	"roll":    true,
	"tube":    true,

	// Generic food terms that don't help narrow down
	"food":    true,
	"item":    true,
	"product": true,
	"brand":   true,
}

// NewQueryPreprocessor creates a new query preprocessor
func NewQueryPreprocessor(enableDebugLogging bool) *QueryPreprocessor {
	return &QueryPreprocessor{
		enableDebugLogging: enableDebugLogging,
	}
}

// PreprocessQuery cleans a product name for USDA API search
// Removes size/quantity info, pack counts, marketing terms, and normalizes whitespace
func (p *QueryPreprocessor) PreprocessQuery(productName, brand string) string {
	if productName == "" {
		return ""
	}

	original := productName

	// Step 1: Remove size/quantity patterns (e.g., "128 fl oz", "1.5 liter")
	cleaned := sizeQuantityPattern.ReplaceAllString(productName, " ")

	// Step 2: Remove pack/count patterns (e.g., "12 pack", "pack of 6")
	cleaned = packCountPattern.ReplaceAllString(cleaned, " ")

	// Step 3: Remove standalone numbers at boundaries
	cleaned = standaloneNumberPattern.ReplaceAllString(cleaned, " ")

	// Step 4: Remove noise words
	cleaned = p.removeNoiseWords(cleaned)

	// Step 5: Clean up punctuation that's now orphaned
	cleaned = cleanOrphanedPunctuation(cleaned)

	// Step 6: Normalize whitespace
	cleaned = multiSpacePattern.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	// Step 7: Prepend brand if provided and not already in the cleaned name (case-insensitive check)
	if brand != "" {
		cleanedLower := strings.ToLower(cleaned)
		brandLower := strings.ToLower(brand)
		if !strings.Contains(cleanedLower, brandLower) {
			cleaned = brand + " " + cleaned
		}
	}

	// Step 8: Limit query length to avoid USDA API issues
	if len(cleaned) > 100 {
		cleaned = cleaned[:100]
		// Try to cut at word boundary
		if lastSpace := strings.LastIndex(cleaned, " "); lastSpace > 50 {
			cleaned = cleaned[:lastSpace]
		}
	}

	if p.enableDebugLogging {
		log.Printf("[PREPROCESS] Input: %q → Output: %q", original, cleaned)
	}

	return cleaned
}

// removeNoiseWords removes marketing and generic terms from the query
func (p *QueryPreprocessor) removeNoiseWords(s string) string {
	words := strings.Fields(strings.ToLower(s))
	var kept []string

	for _, word := range words {
		// Clean punctuation from word for checking
		cleanWord := strings.Trim(word, ",.!?;:-'\"")

		if !queryNoiseWords[cleanWord] {
			// Preserve original word (with punctuation)
			kept = append(kept, word)
		}
	}

	return strings.Join(kept, " ")
}

// cleanOrphanedPunctuation removes punctuation that's now alone (e.g., lone commas)
func cleanOrphanedPunctuation(s string) string {
	// Remove lone punctuation surrounded by spaces
	result := regexp.MustCompile(`\s+[,\-;:]+\s+`).ReplaceAllString(s, " ")
	// Remove trailing punctuation (except periods which might be abbreviations)
	result = regexp.MustCompile(`[,\-;:]+\s*$`).ReplaceAllString(result, "")
	// Remove leading punctuation
	result = regexp.MustCompile(`^\s*[,\-;:]+`).ReplaceAllString(result, "")
	return result
}

// ExtractFoodKeywords extracts the most important food-related keywords from text
// Returns a slice of keywords ordered by importance
func (p *QueryPreprocessor) ExtractFoodKeywords(text string) []string {
	tokens := tokenize(text)

	// Separate into categories by weight
	var highPriority []string  // Food terms (weight 3)
	var medPriority []string   // Descriptive terms (weight 2)
	var lowPriority []string   // Other terms (weight 1)

	for _, token := range tokens {
		if foodTerms[token] {
			highPriority = append(highPriority, token)
		} else if descriptiveTerms[token] {
			medPriority = append(medPriority, token)
		} else {
			lowPriority = append(lowPriority, token)
		}
	}

	// Combine in priority order
	result := make([]string, 0, len(tokens))
	result = append(result, highPriority...)
	result = append(result, medPriority...)
	result = append(result, lowPriority...)

	return result
}
