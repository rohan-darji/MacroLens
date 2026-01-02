import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { NutritionData } from '@/types/nutrition';
import type { ProductInfo } from '@/types/product';

// Mock constants
vi.mock('@/config/constants', () => ({
  API_BASE_URL: 'http://localhost:8080',
  ENDPOINTS: {
    NUTRITION_SEARCH: '/api/v1/nutrition/search',
    HEALTH: '/api/v1/health',
  },
  CACHE_TTL_MS: 7 * 24 * 60 * 60 * 1000, // 7 days in ms
  DEBUG_MODE: false,
}));

describe('API Client', () => {
  let mockFetch: ReturnType<typeof vi.fn>;

  const mockProductInfo: ProductInfo = {
    name: 'Great Value Whole Milk',
    brand: 'Great Value',
    url: 'https://www.walmart.com/ip/test/12345',
    retailer: 'walmart',
  };

  const mockNutritionData: NutritionData = {
    fdcId: 123456,
    productName: 'Whole Milk',
    brand: 'Great Value',
    servingSize: 240,
    servingSizeUnit: 'ml',
    confidence: 85,
    source: 'USDA',
    nutrients: {
      calories: 150,
      protein: 8,
      carbohydrates: 12,
      totalFat: 8,
      saturatedFat: 5,
      fiber: 0,
      sugar: 12,
      sodium: 130,
    },
  };

  beforeEach(() => {
    // Reset modules to clear any cached state
    vi.resetModules();

    // Mock fetch
    mockFetch = vi.fn();
    global.fetch = mockFetch;

    // Mock chrome.storage.local
    global.chrome = {
      storage: {
        local: {
          get: vi.fn().mockResolvedValue({}),
          set: vi.fn().mockResolvedValue(undefined),
          remove: vi.fn().mockResolvedValue(undefined),
        },
      },
    } as any;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('searchNutrition', () => {
    it('should return cached data if available', async () => {
      const cachedData = {
        data: mockNutritionData,
        timestamp: Date.now(),
        cacheKey: 'nutrition:great value whole milk:great value',
      };

      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        'nutrition:great value whole milk:great value': cachedData,
      });

      const { searchNutrition } = await import('@/background/api-client');
      const result = await searchNutrition(mockProductInfo);

      expect(result.cached).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.data?.source).toBe('Cache');
      expect(mockFetch).not.toHaveBeenCalled();
    });

    it('should call API when cache miss', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockNutritionData),
      });

      const { searchNutrition } = await import('@/background/api-client');
      const result = await searchNutrition(mockProductInfo);

      expect(result.cached).toBe(false);
      expect(result.data).toEqual(mockNutritionData);
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/nutrition/search',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
        })
      );
    });

    it('should cache high confidence results', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ ...mockNutritionData, confidence: 90 }),
      });

      const { searchNutrition } = await import('@/background/api-client');
      await searchNutrition(mockProductInfo);

      expect(global.chrome.storage.local.set).toHaveBeenCalled();
    });

    it('should not cache low confidence results', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ ...mockNutritionData, confidence: 50 }),
      });

      const { searchNutrition } = await import('@/background/api-client');
      await searchNutrition(mockProductInfo);

      expect(global.chrome.storage.local.set).not.toHaveBeenCalled();
    });

    it('should return error on API failure', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.resolve({ error: 'Server error' }),
      });

      const { searchNutrition } = await import('@/background/api-client');
      const result = await searchNutrition(mockProductInfo);

      expect(result.error).toBe('Server error');
      expect(result.cached).toBe(false);
      expect(result.data).toBeUndefined();
    });

    it('should return error message with status on API failure without error body', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: () => Promise.reject(new Error('No JSON')),
      });

      const { searchNutrition } = await import('@/background/api-client');
      const result = await searchNutrition(mockProductInfo);

      expect(result.error).toContain('404');
      expect(result.error).toContain('Not Found');
    });

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValue(new Error('Network error'));

      const { searchNutrition } = await import('@/background/api-client');
      const result = await searchNutrition(mockProductInfo);

      expect(result.error).toBe('Network error');
      expect(result.cached).toBe(false);
    });

    it('should remove expired cache entries', async () => {
      const expiredData = {
        data: mockNutritionData,
        timestamp: Date.now() - 8 * 24 * 60 * 60 * 1000, // 8 days ago
        cacheKey: 'nutrition:great value whole milk:great value',
      };

      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        'nutrition:great value whole milk:great value': expiredData,
      });

      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockNutritionData),
      });

      const { searchNutrition } = await import('@/background/api-client');
      await searchNutrition(mockProductInfo);

      expect(global.chrome.storage.local.remove).toHaveBeenCalled();
      expect(mockFetch).toHaveBeenCalled();
    });

    it('should generate correct cache key', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ ...mockNutritionData, confidence: 90 }),
      });

      const { searchNutrition } = await import('@/background/api-client');
      await searchNutrition(mockProductInfo);

      // Check that set was called with the correct cache key format
      const setCall = (global.chrome.storage.local.set as ReturnType<typeof vi.fn>).mock
        .calls[0][0];
      const key = Object.keys(setCall)[0];
      expect(key).toBe('nutrition:great value whole milk:great value');
    });

    it('should handle product without brand', async () => {
      const productWithoutBrand: ProductInfo = {
        name: 'Generic Milk',
        url: 'https://www.walmart.com/ip/test/12345',
        retailer: 'walmart',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ ...mockNutritionData, confidence: 90 }),
      });

      const { searchNutrition } = await import('@/background/api-client');
      await searchNutrition(productWithoutBrand);

      const setCall = (global.chrome.storage.local.set as ReturnType<typeof vi.fn>).mock
        .calls[0][0];
      const key = Object.keys(setCall)[0];
      expect(key).toBe('nutrition:generic milk:');
    });

    it('should normalize cache keys (lowercase, remove special chars)', async () => {
      const productWithSpecialChars: ProductInfo = {
        name: "O'Brien's Super-Food! (100% Natural)",
        brand: 'O\'Brien\'s',
        url: 'https://www.walmart.com/ip/test/12345',
        retailer: 'walmart',
      };

      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ ...mockNutritionData, confidence: 90 }),
      });

      const { searchNutrition } = await import('@/background/api-client');
      await searchNutrition(productWithSpecialChars);

      const setCall = (global.chrome.storage.local.set as ReturnType<typeof vi.fn>).mock
        .calls[0][0];
      const key = Object.keys(setCall)[0];
      // Should be lowercase, no special chars
      expect(key).toBe('nutrition:obriens superfood 100 natural:obriens');
    });
  });

  describe('clearCache', () => {
    it('should clear all nutrition cache entries', async () => {
      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        'nutrition:product1:brand1': {},
        'nutrition:product2:brand2': {},
        'other:key': {},
      });

      const { clearCache } = await import('@/background/api-client');
      await clearCache();

      expect(global.chrome.storage.local.remove).toHaveBeenCalledWith([
        'nutrition:product1:brand1',
        'nutrition:product2:brand2',
      ]);
    });

    it('should not remove non-nutrition keys', async () => {
      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        settings: {},
        preferences: {},
      });

      const { clearCache } = await import('@/background/api-client');
      await clearCache();

      expect(global.chrome.storage.local.remove).not.toHaveBeenCalled();
    });

    it('should handle empty cache', async () => {
      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockResolvedValue({});

      const { clearCache } = await import('@/background/api-client');
      await expect(clearCache()).resolves.not.toThrow();
    });

    it('should throw on storage error', async () => {
      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Storage error')
      );

      const { clearCache } = await import('@/background/api-client');
      await expect(clearCache()).rejects.toThrow('Storage error');
    });
  });

  describe('getCacheStats', () => {
    it('should return cache statistics', async () => {
      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        'nutrition:product1:brand1': {},
        'nutrition:product2:brand2': {},
        'nutrition:product3:': {},
        'other:key': {},
      });

      const { getCacheStats } = await import('@/background/api-client');
      const stats = await getCacheStats();

      expect(stats.count).toBe(3);
      expect(stats.keys).toHaveLength(3);
      expect(stats.keys).toContain('nutrition:product1:brand1');
    });

    it('should return empty stats for empty cache', async () => {
      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockResolvedValue({});

      const { getCacheStats } = await import('@/background/api-client');
      const stats = await getCacheStats();

      expect(stats.count).toBe(0);
      expect(stats.keys).toHaveLength(0);
    });

    it('should handle storage errors gracefully', async () => {
      (global.chrome.storage.local.get as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Storage error')
      );

      const { getCacheStats } = await import('@/background/api-client');
      const stats = await getCacheStats();

      expect(stats.count).toBe(0);
      expect(stats.keys).toHaveLength(0);
    });
  });
});
