package usda

import (
	"fmt"

	"github.com/macrolens/backend/internal/domain"
)

// USDA Nutrient IDs for key macronutrients
const (
	NutrientIDEnergy       = 1008 // Calories (kcal)
	NutrientIDProtein      = 1003 // Protein (g)
	NutrientIDCarbohydrate = 1005 // Carbohydrates (g)
	NutrientIDTotalFat     = 1004 // Total Fat (g)
)

// MapToNutritionData converts USDA food data to our domain NutritionData model
func MapToNutritionData(usdaFood *domain.USDAFood, confidence float64) *domain.NutritionData {
	nutrients := extractNutrients(usdaFood.Nutrients)

	return &domain.NutritionData{
		FdcID:           fmt.Sprintf("%d", usdaFood.FdcID),
		ProductName:     usdaFood.Description,
		ServingSize:     "100", // USDA typically uses 100g as standard serving
		ServingSizeUnit: "g",
		Nutrients:       nutrients,
		Confidence:      confidence,
		Source:          "USDA",
	}
}

// extractNutrients extracts the key macronutrients from USDA nutrient list
func extractNutrients(usdaNutrients []domain.USDANutrient) domain.Nutrients {
	nutrients := domain.Nutrients{}

	for _, nutrient := range usdaNutrients {
		switch nutrient.NutrientID {
		case NutrientIDEnergy:
			nutrients.Calories = nutrient.Value
		case NutrientIDProtein:
			nutrients.Protein = nutrient.Value
		case NutrientIDCarbohydrate:
			nutrients.Carbohydrates = nutrient.Value
		case NutrientIDTotalFat:
			nutrients.TotalFat = nutrient.Value
		}
	}

	return nutrients
}

// FindNutrientValue finds a specific nutrient value by ID
func FindNutrientValue(nutrients []domain.USDANutrient, nutrientID int) float64 {
	for _, nutrient := range nutrients {
		if nutrient.NutrientID == nutrientID {
			return nutrient.Value
		}
	}
	return 0.0
}
