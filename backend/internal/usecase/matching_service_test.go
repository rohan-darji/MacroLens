package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/macrolens/backend/internal/domain"
)

func TestNewMatchingService(t *testing.T) {
	t.Run("creates service with provided threshold", func(t *testing.T) {
		svc := NewMatchingService(MatchConfig{MinConfidenceThreshold: 50})
		if svc.minConfidenceThreshold != 50 {
			t.Errorf("minConfidenceThreshold = %v, want 50", svc.minConfidenceThreshold)
		}
	})

	t.Run("uses default threshold when zero", func(t *testing.T) {
		svc := NewMatchingService(MatchConfig{MinConfidenceThreshold: 0})
		if svc.minConfidenceThreshold != 40 {
			t.Errorf("minConfidenceThreshold = %v, want 40 (default)", svc.minConfidenceThreshold)
		}
	})

	t.Run("uses default threshold when negative", func(t *testing.T) {
		svc := NewMatchingService(MatchConfig{MinConfidenceThreshold: -10})
		if svc.minConfidenceThreshold != 40 {
			t.Errorf("minConfidenceThreshold = %v, want 40 (default)", svc.minConfidenceThreshold)
		}
	})
}

func TestFindBestMatch(t *testing.T) {
	svc := NewMatchingService(MatchConfig{MinConfidenceThreshold: 40})
	ctx := context.Background()

	t.Run("returns error for nil request", func(t *testing.T) {
		_, err := svc.FindBestMatch(ctx, nil, []domain.USDAFood{})
		if !errors.Is(err, domain.ErrInvalidRequest) {
			t.Errorf("error = %v, want ErrInvalidRequest", err)
		}
	})

	t.Run("returns error for empty product name", func(t *testing.T) {
		request := &domain.SearchRequest{ProductName: ""}
		_, err := svc.FindBestMatch(ctx, request, []domain.USDAFood{})
		if !errors.Is(err, domain.ErrInvalidRequest) {
			t.Errorf("error = %v, want ErrInvalidRequest", err)
		}
	})

	t.Run("returns error for empty foods list", func(t *testing.T) {
		request := &domain.SearchRequest{ProductName: "whole milk"}
		_, err := svc.FindBestMatch(ctx, request, []domain.USDAFood{})
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Errorf("error = %v, want ErrProductNotFound", err)
		}
	})

	t.Run("finds exact match with high confidence", func(t *testing.T) {
		request := &domain.SearchRequest{ProductName: "whole milk"}
		foods := []domain.USDAFood{
			{FdcID: 123, Description: "Whole Milk"},
			{FdcID: 456, Description: "Skim Milk"},
		}

		result, err := svc.FindBestMatch(ctx, request, foods)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.FdcID != "123" {
			t.Errorf("FdcID = %v, want 123", result.FdcID)
		}
		if result.MatchScore < 40 {
			t.Errorf("MatchScore = %v, want >= 40", result.MatchScore)
		}
	})

	t.Run("applies brand bonus when brand matches", func(t *testing.T) {
		requestWithBrand := &domain.SearchRequest{
			ProductName: "whole milk",
			Brand:       "Great Value",
		}
		requestWithoutBrand := &domain.SearchRequest{
			ProductName: "whole milk",
		}
		foods := []domain.USDAFood{
			{FdcID: 123, Description: "Great Value Whole Milk"},
		}

		resultWithBrand, err := svc.FindBestMatch(ctx, requestWithBrand, foods)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resultWithoutBrand, err := svc.FindBestMatch(ctx, requestWithoutBrand, foods)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Brand match should add ~25 points, but may vary slightly due to weighted scoring
		// With brand bonus constant at 25.0, we expect close to that
		scoreDiff := resultWithBrand.MatchScore - resultWithoutBrand.MatchScore
		if scoreDiff < 15 || scoreDiff > 30 {
			t.Errorf("Brand bonus = %v, want between 15 and 30", scoreDiff)
		}
	})

	t.Run("returns low confidence error for poor match", func(t *testing.T) {
		svc := NewMatchingService(MatchConfig{MinConfidenceThreshold: 80})
		request := &domain.SearchRequest{ProductName: "chocolate cake"}
		foods := []domain.USDAFood{
			{FdcID: 123, Description: "Grilled Chicken Breast"},
		}

		result, err := svc.FindBestMatch(ctx, request, foods)
		if !errors.Is(err, domain.ErrLowConfidence) {
			t.Errorf("error = %v, want ErrLowConfidence", err)
		}
		if result == nil {
			t.Error("expected result to be returned even with low confidence")
		}
	})

	t.Run("selects best match from multiple options", func(t *testing.T) {
		request := &domain.SearchRequest{ProductName: "whole milk gallon"}
		foods := []domain.USDAFood{
			{FdcID: 111, Description: "Skim Milk"},
			{FdcID: 222, Description: "Whole Milk, Gallon"},
			{FdcID: 333, Description: "Chocolate Milk"},
		}

		result, err := svc.FindBestMatch(ctx, request, foods)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.FdcID != "222" {
			t.Errorf("FdcID = %v, want 222 (best match)", result.FdcID)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		request := &domain.SearchRequest{ProductName: "milk"}
		foods := []domain.USDAFood{
			{FdcID: 123, Description: "Whole Milk"},
		}

		_, err := svc.FindBestMatch(ctx, request, foods)
		if err == nil {
			t.Error("expected context cancellation error")
		}
	})

	t.Run("returns matched tokens", func(t *testing.T) {
		request := &domain.SearchRequest{ProductName: "whole milk vitamin d"}
		foods := []domain.USDAFood{
			{FdcID: 123, Description: "Whole Milk with Vitamin D added"},
		}

		result, err := svc.FindBestMatch(ctx, request, foods)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.MatchedTokens) == 0 {
			t.Error("expected matched tokens to be populated")
		}
		// Check that common tokens are in the matched list
		hasWhole := false
		hasMilk := false
		for _, token := range result.MatchedTokens {
			if token == "whole" {
				hasWhole = true
			}
			if token == "milk" {
				hasMilk = true
			}
		}
		if !hasWhole || !hasMilk {
			t.Errorf("MatchedTokens = %v, want to include 'whole' and 'milk'", result.MatchedTokens)
		}
	})
}

func TestTokenize(t *testing.T) {
	t.Run("converts to lowercase", func(t *testing.T) {
		tokens := tokenize("WHOLE MILK")
		for _, token := range tokens {
			if token != "whole" && token != "milk" {
				t.Errorf("unexpected token: %v", token)
			}
		}
	})

	t.Run("removes punctuation", func(t *testing.T) {
		tokens := tokenize("milk, whole (vitamin d)")
		for _, token := range tokens {
			if token == "," || token == "(" || token == ")" {
				t.Errorf("punctuation should be removed: %v", token)
			}
		}
	})

	t.Run("filters stop words", func(t *testing.T) {
		tokens := tokenize("milk with vitamin a and d")
		stopWords := map[string]bool{"with": true, "and": true, "a": true}
		for _, token := range tokens {
			if stopWords[token] {
				t.Errorf("stop word should be filtered: %v", token)
			}
		}
	})

	t.Run("returns empty slice for empty string", func(t *testing.T) {
		tokens := tokenize("")
		if len(tokens) != 0 {
			t.Errorf("expected empty slice, got %v", tokens)
		}
	})

	t.Run("handles special characters", func(t *testing.T) {
		tokens := tokenize("2% milk - reduced fat")
		// Should contain meaningful tokens
		found := false
		for _, token := range tokens {
			if token == "milk" || token == "reduced" || token == "fat" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected meaningful tokens, got %v", tokens)
		}
	})
}

func TestCalculateMatchScore(t *testing.T) {
	svc := NewMatchingService(MatchConfig{MinConfidenceThreshold: 40})

	t.Run("returns high score for identical strings after bonuses", func(t *testing.T) {
		score, _ := svc.calculateMatchScore("whole milk", "", "whole milk", "")
		// With weighted scoring: milk (3.0) + whole (2.0) = 5.0 total weight
		// 100% match = 70 base + 10 substring bonus = 80+
		if score < 70 {
			t.Errorf("score = %v, want >= 70 for identical strings", score)
		}
	})

	t.Run("returns 0 for completely different strings", func(t *testing.T) {
		score, _ := svc.calculateMatchScore("chocolate cake", "", "grilled salmon", "")
		if score > 20 {
			t.Errorf("score = %v, want < 20 for unrelated items", score)
		}
	})

	t.Run("returns partial score for partial match", func(t *testing.T) {
		score, _ := svc.calculateMatchScore("whole milk", "", "whole milk reduced fat", "")
		if score < 40 || score > 100 {
			t.Errorf("score = %v, want between 40 and 100", score)
		}
	})

	t.Run("handles empty product name", func(t *testing.T) {
		score, matched := svc.calculateMatchScore("", "", "whole milk", "")
		if score != 0 {
			t.Errorf("score = %v, want 0", score)
		}
		if len(matched) != 0 {
			t.Errorf("matched = %v, want empty", matched)
		}
	})

	t.Run("handles empty USDA description", func(t *testing.T) {
		score, matched := svc.calculateMatchScore("whole milk", "", "", "")
		if score != 0 {
			t.Errorf("score = %v, want 0", score)
		}
		if len(matched) != 0 {
			t.Errorf("matched = %v, want empty", matched)
		}
	})

	t.Run("applies data type bonus for Branded", func(t *testing.T) {
		scoreBranded, _ := svc.calculateMatchScore("whole milk", "", "whole milk", "Branded")
		scoreNoType, _ := svc.calculateMatchScore("whole milk", "", "whole milk", "")
		// Branded should add 10 points
		diff := scoreBranded - scoreNoType
		if diff < 9 || diff > 11 {
			t.Errorf("Branded bonus = %v, want approximately 10", diff)
		}
	})

	t.Run("applies data type bonus for Survey", func(t *testing.T) {
		scoreSurvey, _ := svc.calculateMatchScore("whole milk", "", "whole milk", "Survey (FNDDS)")
		scoreNoType, _ := svc.calculateMatchScore("whole milk", "", "whole milk", "")
		// Survey should add 5 points
		diff := scoreSurvey - scoreNoType
		if diff < 4 || diff > 6 {
			t.Errorf("Survey bonus = %v, want approximately 5", diff)
		}
	})

	t.Run("applies data type bonus for Foundation", func(t *testing.T) {
		scoreFoundation, _ := svc.calculateMatchScore("whole milk", "", "whole milk", "Foundation")
		scoreNoType, _ := svc.calculateMatchScore("whole milk", "", "whole milk", "")
		// Foundation should add 3 points
		diff := scoreFoundation - scoreNoType
		if diff < 2 || diff > 4 {
			t.Errorf("Foundation bonus = %v, want approximately 3", diff)
		}
	})
}

func TestFindIntersection(t *testing.T) {
	t.Run("finds common tokens", func(t *testing.T) {
		tokens1 := []string{"whole", "milk", "vitamin"}
		tokens2 := []string{"milk", "whole", "fat"}
		count, matched := findIntersection(tokens1, tokens2)
		if count != 2 {
			t.Errorf("count = %v, want 2", count)
		}
		if len(matched) != 2 {
			t.Errorf("matched length = %v, want 2", len(matched))
		}
	})

	t.Run("returns zero for no overlap", func(t *testing.T) {
		tokens1 := []string{"chocolate", "cake"}
		tokens2 := []string{"grilled", "chicken"}
		count, matched := findIntersection(tokens1, tokens2)
		if count != 0 {
			t.Errorf("count = %v, want 0", count)
		}
		if len(matched) != 0 {
			t.Errorf("matched length = %v, want 0", len(matched))
		}
	})

	t.Run("handles empty slices", func(t *testing.T) {
		count, matched := findIntersection([]string{}, []string{"milk"})
		if count != 0 || len(matched) != 0 {
			t.Error("expected empty result for empty input")
		}
	})
}

func TestFindUnion(t *testing.T) {
	t.Run("counts unique tokens", func(t *testing.T) {
		tokens1 := []string{"whole", "milk"}
		tokens2 := []string{"milk", "fat"}
		count := findUnion(tokens1, tokens2)
		if count != 3 {
			t.Errorf("count = %v, want 3 (whole, milk, fat)", count)
		}
	})

	t.Run("handles duplicates within same slice", func(t *testing.T) {
		tokens1 := []string{"milk", "milk", "milk"}
		tokens2 := []string{"milk"}
		count := findUnion(tokens1, tokens2)
		if count != 1 {
			t.Errorf("count = %v, want 1", count)
		}
	})

	t.Run("handles empty slices", func(t *testing.T) {
		count := findUnion([]string{}, []string{})
		if count != 0 {
			t.Errorf("count = %v, want 0", count)
		}
	})
}

func TestTokenizeWithWeights(t *testing.T) {
	t.Run("assigns high weight to food terms", func(t *testing.T) {
		tokens := tokenizeWithWeights("chicken breast milk")
		for _, tw := range tokens {
			if tw.Token == "chicken" || tw.Token == "milk" {
				if tw.Weight != weightFood {
					t.Errorf("token %q weight = %v, want %v (food)", tw.Token, tw.Weight, weightFood)
				}
			}
		}
	})

	t.Run("assigns medium weight to descriptive terms", func(t *testing.T) {
		tokens := tokenizeWithWeights("whole organic fresh")
		for _, tw := range tokens {
			if tw.Token == "whole" || tw.Token == "organic" || tw.Token == "fresh" {
				if tw.Weight != weightDescriptive {
					t.Errorf("token %q weight = %v, want %v (descriptive)", tw.Token, tw.Weight, weightDescriptive)
				}
			}
		}
	})

	t.Run("assigns default weight to other terms", func(t *testing.T) {
		tokens := tokenizeWithWeights("premium deluxe")
		for _, tw := range tokens {
			// These aren't food or descriptive terms, should get default weight
			if tw.Weight != weightDefault {
				t.Errorf("token %q weight = %v, want %v (default)", tw.Token, tw.Weight, weightDefault)
			}
		}
	})
}

func TestGetTokenWeight(t *testing.T) {
	testCases := []struct {
		token string
		want  float64
	}{
		{"milk", weightFood},
		{"chicken", weightFood},
		{"bread", weightFood},
		{"whole", weightDescriptive},
		{"organic", weightDescriptive},
		{"fresh", weightDescriptive},
		{"xyz", weightDefault},
		{"abc", weightDefault},
	}

	for _, tc := range testCases {
		t.Run(tc.token, func(t *testing.T) {
			got := getTokenWeight(tc.token)
			if got != tc.want {
				t.Errorf("getTokenWeight(%q) = %v, want %v", tc.token, got, tc.want)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	testCases := []struct {
		input string
		want  bool
	}{
		{"123", true},
		{"12", true},
		{"0", true},
		{"", false},
		{"12a", false},
		{"abc", false},
		{"12.5", false}, // dot is not a digit
		{"12 34", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := isNumeric(tc.input)
			if got != tc.want {
				t.Errorf("isNumeric(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	testCases := []struct {
		s1   string
		s2   string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},       // substitution
		{"abc", "abcd", 1},      // insertion
		{"abcd", "abc", 1},      // deletion
		{"kitten", "sitting", 3}, // classic example
		{"milk", "mlik", 2},     // transposition (2 edits)
		{"chicken", "chiken", 1}, // missing letter
	}

	for _, tc := range testCases {
		t.Run(tc.s1+"_"+tc.s2, func(t *testing.T) {
			got := levenshteinDistance(tc.s1, tc.s2)
			if got != tc.want {
				t.Errorf("levenshteinDistance(%q, %q) = %v, want %v", tc.s1, tc.s2, got, tc.want)
			}
		})
	}
}

func TestFuzzyTokenMatch(t *testing.T) {
	testCases := []struct {
		token1    string
		token2    string
		threshold int
		want      bool
	}{
		{"milk", "milk", 1, true},            // identical
		{"milk", "mlik", 1, false},           // short token, fuzzy disabled
		{"chicken", "chiken", 1, true},       // edit distance 1
		{"chicken", "chickn", 1, true},       // edit distance 1
		{"chicken", "chikin", 1, false},      // edit distance 2
		{"chicken", "chikin", 2, true},       // within threshold 2
		{"abc", "abd", 1, false},             // too short for fuzzy
		{"abcde", "abcdf", 1, true},          // 5 chars, edit distance 1
		{"organic", "organc", 1, true},       // typo (missing i)
		{"strawberry", "strawbery", 1, true}, // missing letter
	}

	for _, tc := range testCases {
		t.Run(tc.token1+"_"+tc.token2, func(t *testing.T) {
			got := fuzzyTokenMatch(tc.token1, tc.token2, tc.threshold)
			if got != tc.want {
				t.Errorf("fuzzyTokenMatch(%q, %q, %d) = %v, want %v",
					tc.token1, tc.token2, tc.threshold, got, tc.want)
			}
		})
	}
}

func TestFuzzyMatchingEnabled(t *testing.T) {
	t.Run("fuzzy matching finds close matches when enabled", func(t *testing.T) {
		svc := NewMatchingService(MatchConfig{
			MinConfidenceThreshold: 40,
			EnableFuzzyMatching:    true,
			FuzzyEditDistance:      1,
		})
		ctx := context.Background()

		// "chiken" is a typo for "chicken"
		request := &domain.SearchRequest{ProductName: "grilled chiken breast"}
		foods := []domain.USDAFood{
			{FdcID: 123, Description: "Grilled Chicken Breast", DataType: "Foundation"},
		}

		result, err := svc.FindBestMatch(ctx, request, foods)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that fuzzy match token is in matched tokens
		hasFuzzy := false
		for _, token := range result.MatchedTokens {
			if token == "chiken~chicken" {
				hasFuzzy = true
				break
			}
		}
		if !hasFuzzy {
			t.Errorf("expected fuzzy match token 'chiken~chicken' in %v", result.MatchedTokens)
		}
	})

	t.Run("fuzzy matching disabled does not find typos", func(t *testing.T) {
		svc := NewMatchingService(MatchConfig{
			MinConfidenceThreshold: 40,
			EnableFuzzyMatching:    false,
		})
		ctx := context.Background()

		request := &domain.SearchRequest{ProductName: "grilled chiken breast"}
		foods := []domain.USDAFood{
			{FdcID: 123, Description: "Grilled Chicken Breast", DataType: "Foundation"},
		}

		result, _ := svc.FindBestMatch(ctx, request, foods)

		// Check that there's no fuzzy match token
		for _, token := range result.MatchedTokens {
			if token == "chiken~chicken" {
				t.Error("fuzzy match should not occur when disabled")
			}
		}
	})
}

func TestTokenizeFiltersNumericTokens(t *testing.T) {
	t.Run("filters pure numeric tokens", func(t *testing.T) {
		tokens := tokenize("milk 128 fl oz 12 pack")
		for _, token := range tokens {
			if isNumeric(token) {
				t.Errorf("numeric token should be filtered: %v", token)
			}
		}
		// "fl" and "oz" should also be filtered as stop words
		found := make(map[string]bool)
		for _, token := range tokens {
			found[token] = true
		}
		if found["fl"] || found["oz"] || found["pack"] {
			t.Errorf("stop words should be filtered, got tokens: %v", tokens)
		}
		if !found["milk"] {
			t.Errorf("'milk' should be kept, got tokens: %v", tokens)
		}
	})
}

func TestRealisticWalmartProducts(t *testing.T) {
	svc := NewMatchingService(MatchConfig{
		MinConfidenceThreshold: 40,
		EnableFuzzyMatching:    true,
		FuzzyEditDistance:      1,
	})
	ctx := context.Background()

	testCases := []struct {
		name          string
		productName   string
		brand         string
		usdaFoods     []domain.USDAFood
		wantFdcID     string
		minConfidence float64
	}{
		{
			name:        "whole milk matches correctly",
			productName: "Whole Milk, Vitamin D, Gallon, 128 fl oz",
			brand:       "Great Value",
			usdaFoods: []domain.USDAFood{
				{FdcID: 111, Description: "Skim Milk", DataType: "Foundation"},
				{FdcID: 222, Description: "Great Value Whole Milk, Vitamin D", DataType: "Branded"},
				{FdcID: 333, Description: "Chocolate Milk", DataType: "Foundation"},
			},
			wantFdcID:     "222",
			minConfidence: 50,
		},
		{
			name:        "chicken breast matches",
			productName: "Boneless Skinless Chicken Breasts, 2.5 lb",
			brand:       "Tyson",
			usdaFoods: []domain.USDAFood{
				{FdcID: 111, Description: "Tyson Boneless Skinless Chicken Breast", DataType: "Branded"},
				{FdcID: 222, Description: "Chicken Wings", DataType: "Foundation"},
				{FdcID: 333, Description: "Ground Beef", DataType: "Foundation"},
			},
			wantFdcID:     "111",
			minConfidence: 50,
		},
		{
			name:        "cereal matches",
			productName: "Cheerios Heart Healthy Cereal, 18 oz",
			brand:       "General Mills",
			usdaFoods: []domain.USDAFood{
				{FdcID: 111, Description: "Corn Flakes", DataType: "Foundation"},
				{FdcID: 222, Description: "Cheerios, Whole Grain Oat Cereal", DataType: "Branded"},
				{FdcID: 333, Description: "Oatmeal", DataType: "Foundation"},
			},
			wantFdcID:     "222",
			minConfidence: 40,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &domain.SearchRequest{
				ProductName: tc.productName,
				Brand:       tc.brand,
			}

			result, err := svc.FindBestMatch(ctx, request, tc.usdaFoods)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.FdcID != tc.wantFdcID {
				t.Errorf("FdcID = %v, want %v", result.FdcID, tc.wantFdcID)
			}

			if result.MatchScore < tc.minConfidence {
				t.Errorf("MatchScore = %v, want >= %v", result.MatchScore, tc.minConfidence)
			}
		})
	}
}
