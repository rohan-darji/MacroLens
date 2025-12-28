// nutrition.ts - Nutrition data types

export interface NutritionData {
  fdcId: string;
  productName: string;
  servingSize: string;
  servingSizeUnit: string;
  nutrients: Nutrients;
  confidence: number; // Match confidence 0-100
  source: 'USDA' | 'Cache';
}

export interface Nutrients {
  calories: number;
  protein: number;       // grams
  carbohydrates: number; // grams
  totalFat: number;      // grams
}

export interface CachedNutrition {
  data: NutritionData;
  timestamp: number;
  cacheKey: string;
}
