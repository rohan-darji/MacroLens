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

		// Brand match should add 20 points
		scoreDiff := resultWithBrand.MatchScore - resultWithoutBrand.MatchScore
		if scoreDiff < 19 || scoreDiff > 21 {
			t.Errorf("Brand bonus = %v, want approximately 20", scoreDiff)
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

	t.Run("returns 100 for identical strings after bonuses", func(t *testing.T) {
		score, _ := svc.calculateMatchScore("whole milk", "", "whole milk")
		// Jaccard = 1.0 (100%) + substring bonus (15) = capped at 100
		if score != 100 {
			t.Errorf("score = %v, want 100", score)
		}
	})

	t.Run("returns 0 for completely different strings", func(t *testing.T) {
		score, _ := svc.calculateMatchScore("chocolate cake", "", "grilled salmon")
		if score > 20 {
			t.Errorf("score = %v, want < 20 for unrelated items", score)
		}
	})

	t.Run("returns partial score for partial match", func(t *testing.T) {
		score, _ := svc.calculateMatchScore("whole milk", "", "whole milk reduced fat")
		if score < 40 || score > 100 {
			t.Errorf("score = %v, want between 40 and 100", score)
		}
	})

	t.Run("handles empty product name", func(t *testing.T) {
		score, matched := svc.calculateMatchScore("", "", "whole milk")
		if score != 0 {
			t.Errorf("score = %v, want 0", score)
		}
		if len(matched) != 0 {
			t.Errorf("matched = %v, want empty", matched)
		}
	})

	t.Run("handles empty USDA description", func(t *testing.T) {
		score, matched := svc.calculateMatchScore("whole milk", "", "")
		if score != 0 {
			t.Errorf("score = %v, want 0", score)
		}
		if len(matched) != 0 {
			t.Errorf("matched = %v, want empty", matched)
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
