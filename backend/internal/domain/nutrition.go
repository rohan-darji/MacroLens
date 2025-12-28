package domain

import "time"

// NutritionData represents the complete nutrition information for a food product
type NutritionData struct {
	FdcID           string    `json:"fdcId"`
	ProductName     string    `json:"productName"`
	ServingSize     string    `json:"servingSize"`
	ServingSizeUnit string    `json:"servingSizeUnit"`
	Nutrients       Nutrients `json:"nutrients"`
	Confidence      float64   `json:"confidence"` // Match confidence score 0-100
	Source          string    `json:"source"`     // "USDA" or "Cache"
	CachedAt        time.Time `json:"cachedAt,omitempty"`
}

// Nutrients contains the key macronutrients for MVP
type Nutrients struct {
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`       // grams
	Carbohydrates float64 `json:"carbohydrates"` // grams
	TotalFat      float64 `json:"totalFat"`      // grams
}

// SearchRequest represents a nutrition search request
type SearchRequest struct {
	ProductName string `json:"productName" binding:"required"`
	Brand       string `json:"brand,omitempty"`
	Size        string `json:"size,omitempty"`
}

// USDAFood represents a food item from the USDA FoodData Central API
type USDAFood struct {
	FdcID       string        `json:"fdcId"`
	Description string        `json:"description"`
	DataType    string        `json:"dataType"`
	FoodClass   string        `json:"foodClass,omitempty"`
	Nutrients   []USDANutrient `json:"foodNutrients"`
}

// USDANutrient represents a single nutrient from USDA data
type USDANutrient struct {
	NutrientID     int     `json:"nutrientId"`
	NutrientName   string  `json:"nutrientName"`
	NutrientNumber string  `json:"nutrientNumber,omitempty"`
	UnitName       string  `json:"unitName"`
	Value          float64 `json:"value"`
}

// USDASearchResponse represents the response from USDA search API
type USDASearchResponse struct {
	Foods      []USDAFood `json:"foods"`
	TotalHits  int        `json:"totalHits"`
	CurrentPage int       `json:"currentPage"`
	TotalPages int        `json:"totalPages"`
}
