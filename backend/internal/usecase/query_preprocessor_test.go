package usecase

import (
	"testing"
)

func TestNewQueryPreprocessor(t *testing.T) {
	t.Run("creates preprocessor with debug logging disabled", func(t *testing.T) {
		p := NewQueryPreprocessor(false)
		if p.enableDebugLogging {
			t.Error("expected debug logging to be disabled")
		}
	})

	t.Run("creates preprocessor with debug logging enabled", func(t *testing.T) {
		p := NewQueryPreprocessor(true)
		if !p.enableDebugLogging {
			t.Error("expected debug logging to be enabled")
		}
	})
}

func TestPreprocessQuery(t *testing.T) {
	p := NewQueryPreprocessor(false)

	testCases := []struct {
		name        string
		productName string
		brand       string
		want        string
	}{
		{
			name:        "removes size in fl oz",
			productName: "Coca-Cola, 12 fl oz",
			brand:       "",
			want:        "coca-cola",
		},
		{
			name:        "removes size in oz",
			productName: "Cheerios Cereal, 18 oz",
			brand:       "",
			want:        "cheerios cereal",
		},
		{
			name:        "removes gallon size",
			productName: "Whole Milk, Vitamin D, Gallon, 128 fl oz",
			brand:       "",
			want:        "whole milk, vitamin d, gallon", // gallon as standalone word isn't removed
		},
		{
			name:        "removes pack count",
			productName: "Coca-Cola Soda Pop, 6 pack",
			brand:       "",
			want:        "coca-cola soda pop", // 6 pack should be removed
		},
		{
			name:        "removes lb weight",
			productName: "Tyson Chicken Breasts, 2.5 lb",
			brand:       "",
			want:        "tyson chicken breasts",
		},
		{
			name:        "prepends brand when not in name",
			productName: "Whole Milk, Vitamin D",
			brand:       "Great Value",
			want:        "Great Value whole milk, vitamin d",
		},
		{
			name:        "does not duplicate brand when already in name",
			productName: "Great Value Whole Milk, Vitamin D",
			brand:       "Great Value",
			want:        "Great Value whole milk, vitamin d", // brand appears because "great value" is in lowercase product name
		},
		{
			name:        "removes marketing terms",
			productName: "Premium Select Quality Chicken Breast",
			brand:       "",
			want:        "chicken breast",
		},
		{
			name:        "removes packaging terms",
			productName: "Cheese Slices, Box of American Cheese",
			brand:       "",
			want:        "cheese slices, of american cheese", // "box" is a noise word
		},
		{
			name:        "handles complex Walmart product name",
			productName: "Great Value Whole Milk, Vitamin D, Gallon, 128 fl oz",
			brand:       "Great Value",
			want:        "Great Value whole milk, vitamin d, gallon", // brand prepended, sizes removed
		},
		{
			name:        "handles empty product name",
			productName: "",
			brand:       "",
			want:        "",
		},
		{
			name:        "handles only brand provided",
			productName: "",
			brand:       "Coca-Cola",
			want:        "",
		},
		{
			name:        "removes count notation",
			productName: "Eggs, Large, 12 count",
			brand:       "",
			want:        "eggs",
		},
		{
			name:        "removes ct abbreviation",
			productName: "Granola Bars, 6 ct",
			brand:       "",
			want:        "granola bars",
		},
		{
			name:        "removes ml measurement",
			productName: "Yogurt Drink, 500 ml",
			brand:       "",
			want:        "yogurt drink",
		},
		{
			name:        "handles liter measurement",
			productName: "Sparkling Water, 2 liters",
			brand:       "",
			want:        "sparkling water",
		},
		{
			name:        "handles grams measurement",
			productName: "Chocolate Bar, 100 grams",
			brand:       "",
			want:        "chocolate bar",
		},
		{
			name:        "preserves food descriptors",
			productName: "Organic Whole Grain Bread",
			brand:       "",
			want:        "organic whole grain bread",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.PreprocessQuery(tc.productName, tc.brand)
			if got != tc.want {
				t.Errorf("PreprocessQuery(%q, %q) = %q, want %q",
					tc.productName, tc.brand, got, tc.want)
			}
		})
	}
}

func TestPreprocessQuery_LongInput(t *testing.T) {
	p := NewQueryPreprocessor(false)

	// Create a very long product name
	longName := "Super Premium Deluxe Ultimate Organic Natural Fresh Farm Raised Free Range Grass Fed Antibiotic Free Hormone Free Non-GMO Certified Gluten Free Dairy Free Vegan Friendly Heart Healthy Brain Boosting Energy Enhancing Muscle Building Weight Loss Supporting Immune Strengthening Chicken Breast Tenderloin Filet"

	result := p.PreprocessQuery(longName, "")

	if len(result) > 100 {
		t.Errorf("result length = %d, want <= 100", len(result))
	}
}

func TestExtractFoodKeywords(t *testing.T) {
	p := NewQueryPreprocessor(false)

	t.Run("extracts food terms first", func(t *testing.T) {
		keywords := p.ExtractFoodKeywords("whole milk vitamin d gallon")
		if len(keywords) == 0 {
			t.Error("expected keywords to be extracted")
		}
		// "milk" should be first since it's a food term (high priority)
		found := false
		for i, kw := range keywords {
			if kw == "milk" {
				if i > 2 {
					t.Errorf("'milk' should be in first few positions, got position %d", i)
				}
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'milk' to be in keywords")
		}
	})

	t.Run("returns empty for empty input", func(t *testing.T) {
		keywords := p.ExtractFoodKeywords("")
		if len(keywords) != 0 {
			t.Errorf("expected empty slice, got %v", keywords)
		}
	})
}

func TestRemoveNoiseWords(t *testing.T) {
	p := NewQueryPreprocessor(false)

	testCases := []struct {
		input string
		want  string
	}{
		{"value brand chicken", "chicken"},
		{"premium select milk", "milk"},
		{"great value cheese", "cheese"},
		{"family size box cereal", "cereal"},
		{"", ""},
		{"chicken breast", "chicken breast"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := p.removeNoiseWords(tc.input)
			if got != tc.want {
				t.Errorf("removeNoiseWords(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestCleanOrphanedPunctuation(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{"milk , cheese", "milk cheese"},
		{", milk", " milk"}, // leading comma removed but space remains
		{"milk,", "milk"},
		{"milk - cheese", "milk cheese"},
		{"milk", "milk"},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := cleanOrphanedPunctuation(tc.input)
			if got != tc.want {
				t.Errorf("cleanOrphanedPunctuation(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
