import { describe, it, expect } from 'vitest';
import {
  API_BASE_URL,
  DEBUG_MODE,
  CACHE_TTL_DAYS,
  CACHE_TTL_MS,
  ENDPOINTS,
  WALMART,
  UI
} from '../../src/config/constants';

describe('Constants Configuration', () => {
  describe('API Configuration', () => {
    it('should have API_BASE_URL defined', () => {
      expect(API_BASE_URL).toBeDefined();
      expect(typeof API_BASE_URL).toBe('string');
    });

    it('should use default API_BASE_URL when env var not set', () => {
      // In test environment without VITE_API_BASE_URL set, should use default
      expect(API_BASE_URL).toBe('http://localhost:8080');
    });

    it('should have DEBUG_MODE defined as boolean', () => {
      expect(typeof DEBUG_MODE).toBe('boolean');
    });

    it('should have DEBUG_MODE as false by default in test environment', () => {
      // In test environment, DEBUG_MODE should be false unless explicitly set
      expect(DEBUG_MODE).toBe(false);
    });
  });

  describe('Cache Configuration', () => {
    it('should have default cache TTL in days', () => {
      expect(CACHE_TTL_DAYS).toBe(7);
    });

    it('should calculate cache TTL in milliseconds correctly', () => {
      const expectedMs = CACHE_TTL_DAYS * 24 * 60 * 60 * 1000;
      expect(CACHE_TTL_MS).toBe(expectedMs);
      expect(CACHE_TTL_MS).toBe(604800000); // 7 days in ms
    });

    it('should have CACHE_TTL_MS defined', () => {
      expect(CACHE_TTL_MS).toBeDefined();
      expect(typeof CACHE_TTL_MS).toBe('number');
    });

    it('should correctly derive CACHE_TTL_MS from CACHE_TTL_DAYS', () => {
      // Verify the relationship between days and milliseconds
      expect(CACHE_TTL_MS).toBe(CACHE_TTL_DAYS * 24 * 60 * 60 * 1000);
    });
  });

  describe('Walmart Configuration', () => {
    it('should have correct product URL pattern', () => {
      expect(WALMART.PRODUCT_URL_PATTERN).toBeDefined();
      expect(WALMART.PRODUCT_URL_PATTERN).toBeInstanceOf(RegExp);
    });

    it('should match valid Walmart product URLs', () => {
      expect(WALMART.PRODUCT_URL_PATTERN.test('https://www.walmart.com/ip/Product/123')).toBe(true);
      expect(WALMART.PRODUCT_URL_PATTERN.test('https://www.walmart.com/ip/Product-Name/123456')).toBe(true);
    });

    it('should reject non-Walmart URLs', () => {
      expect(WALMART.PRODUCT_URL_PATTERN.test('https://www.amazon.com/ip/Product/123')).toBe(false);
      expect(WALMART.PRODUCT_URL_PATTERN.test('https://www.walmart.com/browse/dairy')).toBe(false);
    });

    it('should have product page indicator', () => {
      expect(WALMART.PRODUCT_PAGE_INDICATOR).toBe('/ip/');
    });

    it('should detect product page indicator in URLs', () => {
      expect('https://www.walmart.com/ip/Product/123'.includes(WALMART.PRODUCT_PAGE_INDICATOR)).toBe(true);
      expect('https://www.walmart.com/browse/dairy'.includes(WALMART.PRODUCT_PAGE_INDICATOR)).toBe(false);
    });
  });

  describe('Endpoints Configuration', () => {
    it('should have health endpoint', () => {
      expect(ENDPOINTS.HEALTH).toBe('/health');
    });

    it('should have nutrition search endpoint', () => {
      expect(ENDPOINTS.NUTRITION_SEARCH).toBe('/api/v1/nutrition/search');
    });

    it('should have API versioning in endpoints', () => {
      expect(ENDPOINTS.NUTRITION_SEARCH).toContain('/v1/');
    });

    it('should have all required endpoints defined', () => {
      expect(ENDPOINTS).toHaveProperty('HEALTH');
      expect(ENDPOINTS).toHaveProperty('NUTRITION_SEARCH');
    });
  });

  describe('UI Configuration', () => {
    it('should have overlay ID', () => {
      expect(UI.OVERLAY_ID).toBe('macrolens-overlay');
    });

    it('should have overlay class', () => {
      expect(UI.OVERLAY_CLASS).toBe('macrolens-overlay');
    });

    it('should use consistent naming for ID and class', () => {
      expect(UI.OVERLAY_ID).toBe(UI.OVERLAY_CLASS);
    });

    it('should have all required UI properties defined', () => {
      expect(UI).toHaveProperty('OVERLAY_ID');
      expect(UI).toHaveProperty('OVERLAY_CLASS');
    });
  });

  describe('URL Construction', () => {
    it('should construct full API URL correctly', () => {
      const fullURL = `${API_BASE_URL}${ENDPOINTS.NUTRITION_SEARCH}`;
      expect(fullURL).toBe('http://localhost:8080/api/v1/nutrition/search');
    });

    it('should construct health check URL correctly', () => {
      const healthURL = `${API_BASE_URL}${ENDPOINTS.HEALTH}`;
      expect(healthURL).toBe('http://localhost:8080/health');
    });

    it('should handle trailing slash in base URL', () => {
      const baseWithSlash = 'http://localhost:8080/';
      const fullURL = `${baseWithSlash.replace(/\/$/, '')}${ENDPOINTS.NUTRITION_SEARCH}`;
      expect(fullURL).toBe('http://localhost:8080/api/v1/nutrition/search');
    });
  });

  describe('Cache TTL Calculations', () => {
    it('should convert days to milliseconds correctly for 1 day', () => {
      const days = 1;
      const ms = days * 24 * 60 * 60 * 1000;
      expect(ms).toBe(86400000);
    });

    it('should convert days to milliseconds correctly for 7 days', () => {
      const days = 7;
      const ms = days * 24 * 60 * 60 * 1000;
      expect(ms).toBe(604800000);
    });

    it('should convert days to milliseconds correctly for 30 days', () => {
      const days = 30;
      const ms = days * 24 * 60 * 60 * 1000;
      expect(ms).toBe(2592000000);
    });
  });
});
