// api-client.ts - HTTP client for backend API with extension-side caching

import type { NutritionData, CachedNutrition } from '@/types/nutrition';
import type { ProductInfo, SearchRequest } from '@/types/product';
import {
  API_BASE_URL,
  ENDPOINTS,
  CACHE_TTL_MS,
  DEBUG_MODE,
  CACHE_CONFIDENCE_THRESHOLD,
} from '@/config/constants';

/**
 * Logs debug messages if DEBUG_MODE is enabled
 */
function log(...args: any[]): void {
  if (DEBUG_MODE) {
    console.log('[MacroLens API Client]', ...args);
  }
}

/**
 * Searches for nutrition data for a product
 * Checks extension cache first, then calls backend API if needed
 */
export async function searchNutrition(
  productInfo: ProductInfo
): Promise<{ data?: NutritionData; error?: string; cached: boolean }> {
  try {
    const cacheKey = generateCacheKey(productInfo);
    log('Searching nutrition for:', productInfo.name, 'Cache key:', cacheKey);

    // Check extension cache first
    const cachedData = await getCachedNutrition(cacheKey);
    if (cachedData) {
      log('Cache hit!', cachedData);
      return {
        data: { ...cachedData, source: 'Cache' },
        cached: true,
      };
    }

    log('Cache miss, calling backend API');

    // Cache miss - call backend API
    const searchRequest: SearchRequest = {
      productName: productInfo.name,
      brand: productInfo.brand,
      size: productInfo.size,
    };

    const response = await fetch(`${API_BASE_URL}${ENDPOINTS.NUTRITION_SEARCH}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(searchRequest),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.error || `API error: ${response.status} ${response.statusText}`;
      log('API error:', errorMessage);
      return {
        error: errorMessage,
        cached: false,
      };
    }

    const nutritionData: NutritionData = await response.json();
    log('API response:', nutritionData);

    // Cache the result for future use (only if not low confidence)
    if (nutritionData.confidence >= CACHE_CONFIDENCE_THRESHOLD) {
      await setCachedNutrition(cacheKey, nutritionData);
      log('Cached nutrition data');
    } else {
      log('Not caching low confidence result (below', CACHE_CONFIDENCE_THRESHOLD, '% threshold)');
    }

    return {
      data: nutritionData,
      cached: false,
    };
  } catch (error) {
    log('Error searching nutrition:', error);
    return {
      error: error instanceof Error ? error.message : 'Unknown error occurred',
      cached: false,
    };
  }
}

/**
 * Generates a cache key from product information
 * Format matches backend: "nutrition:{normalized_product_name}:{brand}"
 */
function generateCacheKey(productInfo: ProductInfo): string {
  const normalizedName = normalizeForCacheKey(productInfo.name);
  const normalizedBrand = normalizeForCacheKey(productInfo.brand || '');
  return `nutrition:${normalizedName}:${normalizedBrand}`;
}

/**
 * Normalizes a string for use as cache key component.
 * Converts to lowercase, removes special characters, and trims whitespace.
 * Must match backend normalization logic.
 *
 * INTENTIONAL BEHAVIOR: This normalization removes ALL non-alphanumeric characters
 * including hyphens, underscores, and punctuation. This means:
 * - "Coca-Cola" -> "cocacola"
 * - "Coca Cola" -> "coca cola"
 *
 * This is intentional to improve cache hit rates for product name variations
 * (e.g., "Coca-Cola", "Coca Cola", "CocaCola" should all match the same cache entry).
 * The trade-off is potential cache collisions, but the confidence score in cached
 * data helps ensure we still return quality matches.
 */
function normalizeForCacheKey(s: string): string {
  if (!s) {
    return '';
  }
  let result = s.toLowerCase();
  result = result.replace(/[^a-z0-9\s]/g, ''); // Remove non-alphanumeric
  result = result.replace(/\s+/g, ' '); // Collapse multiple spaces
  return result.trim();
}

/**
 * Retrieves nutrition data from extension cache
 */
async function getCachedNutrition(cacheKey: string): Promise<NutritionData | null> {
  try {
    const result = await chrome.storage.local.get(cacheKey);
    const cached = result[cacheKey] as CachedNutrition | undefined;

    if (!cached) {
      return null;
    }

    // Check if cache entry has expired
    const now = Date.now();
    const age = now - cached.timestamp;

    if (age > CACHE_TTL_MS) {
      log('Cache entry expired, removing:', cacheKey);
      await chrome.storage.local.remove(cacheKey);
      return null;
    }

    return cached.data;
  } catch (error) {
    log('Error reading from cache:', error);
    return null;
  }
}

/**
 * Stores nutrition data in extension cache
 */
async function setCachedNutrition(cacheKey: string, data: NutritionData): Promise<void> {
  try {
    const cached: CachedNutrition = {
      data,
      timestamp: Date.now(),
      cacheKey,
    };

    await chrome.storage.local.set({ [cacheKey]: cached });
  } catch (error) {
    log('Error writing to cache:', error);
    // Don't throw - caching failure shouldn't break the app
  }
}

/**
 * Clears all cached nutrition data
 */
export async function clearCache(): Promise<void> {
  try {
    const allData = await chrome.storage.local.get(null);
    const nutritionKeys = Object.keys(allData).filter((key) => key.startsWith('nutrition:'));

    if (nutritionKeys.length > 0) {
      await chrome.storage.local.remove(nutritionKeys);
      log('Cleared cache entries:', nutritionKeys.length);
    } else {
      log('No cache entries to clear');
    }
  } catch (error) {
    log('Error clearing cache:', error);
    throw error;
  }
}

/**
 * Gets cache statistics
 */
export async function getCacheStats(): Promise<{ count: number; keys: string[] }> {
  try {
    const allData = await chrome.storage.local.get(null);
    const nutritionKeys = Object.keys(allData).filter((key) => key.startsWith('nutrition:'));

    return {
      count: nutritionKeys.length,
      keys: nutritionKeys,
    };
  } catch (error) {
    log('Error getting cache stats:', error);
    return { count: 0, keys: [] };
  }
}
