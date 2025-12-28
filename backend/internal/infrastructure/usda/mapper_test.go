package usda

import (
	"testing"

	"github.com/macrolens/backend/internal/domain"
)

func TestMapToNutritionData(t *testing.T) {
	tests := []struct {
		name       string
		usdaFood   *domain.USDAFood
		confidence float64
		want       *domain.NutritionData
	}{
		{
			name: "complete food data",
			usdaFood: &domain.USDAFood{
				FdcID:       "12345",
				Description: "Whole Milk",
				DataType:    "Survey (FNDDS)",
				Nutrients: []domain.USDANutrient{
					{NutrientID: NutrientIDEnergy, NutrientName: "Energy", Value: 149.0, UnitName: "kcal"},
					{NutrientID: NutrientIDProtein, NutrientName: "Protein", Value: 7.7, UnitName: "g"},
					{NutrientID: NutrientIDCarbohydrate, NutrientName: "Carbohydrate", Value: 11.7, UnitName: "g"},
					{NutrientID: NutrientIDTotalFat, NutrientName: "Total Fat", Value: 7.9, UnitName: "g"},
				},
			},
			confidence: 92.5,
			want: &domain.NutritionData{
				FdcID:           "12345",
				ProductName:     "Whole Milk",
				ServingSize:     "100",
				ServingSizeUnit: "g",
				Nutrients: domain.Nutrients{
					Calories:      149.0,
					Protein:       7.7,
					Carbohydrates: 11.7,
					TotalFat:      7.9,
				},
				Confidence: 92.5,
				Source:     "USDA",
			},
		},
		{
			name: "missing some nutrients",
			usdaFood: &domain.USDAFood{
				FdcID:       "67890",
				Description: "Apple",
				Nutrients: []domain.USDANutrient{
					{NutrientID: NutrientIDEnergy, Value: 52.0},
					{NutrientID: NutrientIDCarbohydrate, Value: 14.0},
				},
			},
			confidence: 85.0,
			want: &domain.NutritionData{
				FdcID:           "67890",
				ProductName:     "Apple",
				ServingSize:     "100",
				ServingSizeUnit: "g",
				Nutrients: domain.Nutrients{
					Calories:      52.0,
					Protein:       0.0, // Missing, should default to 0
					Carbohydrates: 14.0,
					TotalFat:      0.0, // Missing, should default to 0
				},
				Confidence: 85.0,
				Source:     "USDA",
			},
		},
		{
			name: "no nutrients",
			usdaFood: &domain.USDAFood{
				FdcID:       "11111",
				Description: "Unknown Food",
				Nutrients:   []domain.USDANutrient{},
			},
			confidence: 50.0,
			want: &domain.NutritionData{
				FdcID:           "11111",
				ProductName:     "Unknown Food",
				ServingSize:     "100",
				ServingSizeUnit: "g",
				Nutrients: domain.Nutrients{
					Calories:      0.0,
					Protein:       0.0,
					Carbohydrates: 0.0,
					TotalFat:      0.0,
				},
				Confidence: 50.0,
				Source:     "USDA",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapToNutritionData(tt.usdaFood, tt.confidence)

			if got.FdcID != tt.want.FdcID {
				t.Errorf("FdcID = %v, want %v", got.FdcID, tt.want.FdcID)
			}
			if got.ProductName != tt.want.ProductName {
				t.Errorf("ProductName = %v, want %v", got.ProductName, tt.want.ProductName)
			}
			if got.ServingSize != tt.want.ServingSize {
				t.Errorf("ServingSize = %v, want %v", got.ServingSize, tt.want.ServingSize)
			}
			if got.ServingSizeUnit != tt.want.ServingSizeUnit {
				t.Errorf("ServingSizeUnit = %v, want %v", got.ServingSizeUnit, tt.want.ServingSizeUnit)
			}
			if got.Confidence != tt.want.Confidence {
				t.Errorf("Confidence = %v, want %v", got.Confidence, tt.want.Confidence)
			}
			if got.Source != tt.want.Source {
				t.Errorf("Source = %v, want %v", got.Source, tt.want.Source)
			}

			// Check nutrients
			if got.Nutrients.Calories != tt.want.Nutrients.Calories {
				t.Errorf("Nutrients.Calories = %v, want %v", got.Nutrients.Calories, tt.want.Nutrients.Calories)
			}
			if got.Nutrients.Protein != tt.want.Nutrients.Protein {
				t.Errorf("Nutrients.Protein = %v, want %v", got.Nutrients.Protein, tt.want.Nutrients.Protein)
			}
			if got.Nutrients.Carbohydrates != tt.want.Nutrients.Carbohydrates {
				t.Errorf("Nutrients.Carbohydrates = %v, want %v", got.Nutrients.Carbohydrates, tt.want.Nutrients.Carbohydrates)
			}
			if got.Nutrients.TotalFat != tt.want.Nutrients.TotalFat {
				t.Errorf("Nutrients.TotalFat = %v, want %v", got.Nutrients.TotalFat, tt.want.Nutrients.TotalFat)
			}
		})
	}
}

func TestFindNutrientValue(t *testing.T) {
	nutrients := []domain.USDANutrient{
		{NutrientID: NutrientIDEnergy, Value: 100.0},
		{NutrientID: NutrientIDProtein, Value: 5.0},
		{NutrientID: NutrientIDCarbohydrate, Value: 20.0},
	}

	tests := []struct {
		name       string
		nutrients  []domain.USDANutrient
		nutrientID int
		want       float64
	}{
		{
			name:       "find existing nutrient",
			nutrients:  nutrients,
			nutrientID: NutrientIDProtein,
			want:       5.0,
		},
		{
			name:       "nutrient not found",
			nutrients:  nutrients,
			nutrientID: NutrientIDTotalFat,
			want:       0.0,
		},
		{
			name:       "empty nutrient list",
			nutrients:  []domain.USDANutrient{},
			nutrientID: NutrientIDEnergy,
			want:       0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindNutrientValue(tt.nutrients, tt.nutrientID)
			if got != tt.want {
				t.Errorf("FindNutrientValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
