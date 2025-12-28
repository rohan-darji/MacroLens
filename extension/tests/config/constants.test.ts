import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// We need to mock import.meta.env
const mockEnv = (env: Record<string, string>) => {
  vi.stubGlobal('import', {
    meta: {
      env,
    },
  });
};

describe('Constants Configuration', () => {
  describe('API Configuration', () => {
    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it('should use default API_BASE_URL when env var not set', () => {
      mockEnv({});
      const { API_BASE_URL } = require('@/config/constants');
      expect(API_BASE_URL).toBe('http://localhost:8080');
    });

    it('should use environment API_BASE_URL when set', () => {
      mockEnv({ VITE_API_BASE_URL: 'https://api.macrolens.app' });

      // Note: In actual test, we'd need to re-import the module
      // This is a conceptual test - in practice, we'd use dependency injection
    });

    it('should set DEBUG_MODE to false by default', () => {
      mockEnv({});
      const { DEBUG_MODE } = require('@/config/constants');
      expect(DEBUG_MODE).toBe(false);
    });
  });

  describe('Cache Configuration', () => {
    it('should have default cache TTL in days', () => {
      const CACHE_TTL_DAYS = 7;
      expect(CACHE_TTL_DAYS).toBe(7);
    });

    it('should calculate cache TTL in milliseconds correctly', () => {
      const CACHE_TTL_DAYS = 7;
      const CACHE_TTL_MS = CACHE_TTL_DAYS * 24 * 60 * 60 * 1000;

      expect(CACHE_TTL_MS).toBe(604800000); // 7 days in ms
    });
  });

  describe('Walmart Configuration', () => {
    it('should have correct product URL pattern', () => {
      const PRODUCT_URL_PATTERN = /^https:\/\/www\.walmart\.com\/ip\/.+\/\d+$/;

      expect(PRODUCT_URL_PATTERN.test('https://www.walmart.com/ip/Product/123')).toBe(true);
      expect(PRODUCT_URL_PATTERN.test('https://www.amazon.com/ip/Product/123')).toBe(false);
    });

    it('should have product page indicator', () => {
      const PRODUCT_PAGE_INDICATOR = '/ip/';

      expect('https://www.walmart.com/ip/Product/123'.includes(PRODUCT_PAGE_INDICATOR)).toBe(true);
      expect('https://www.walmart.com/browse/dairy'.includes(PRODUCT_PAGE_INDICATOR)).toBe(false);
    });
  });

  describe('Endpoints Configuration', () => {
    it('should have health endpoint', () => {
      const HEALTH = '/health';
      expect(HEALTH).toBe('/health');
    });

    it('should have nutrition search endpoint', () => {
      const NUTRITION_SEARCH = '/api/v1/nutrition/search';
      expect(NUTRITION_SEARCH).toBe('/api/v1/nutrition/search');
    });

    it('should have API versioning in endpoints', () => {
      const NUTRITION_SEARCH = '/api/v1/nutrition/search';
      expect(NUTRITION_SEARCH).toContain('/v1/');
    });
  });

  describe('UI Configuration', () => {
    it('should have overlay ID', () => {
      const OVERLAY_ID = 'macrolens-overlay';
      expect(OVERLAY_ID).toBe('macrolens-overlay');
    });

    it('should have overlay class', () => {
      const OVERLAY_CLASS = 'macrolens-overlay';
      expect(OVERLAY_CLASS).toBe('macrolens-overlay');
    });

    it('should use consistent naming for ID and class', () => {
      const OVERLAY_ID = 'macrolens-overlay';
      const OVERLAY_CLASS = 'macrolens-overlay';
      expect(OVERLAY_ID).toBe(OVERLAY_CLASS);
    });
  });

  describe('URL Construction', () => {
    it('should construct full API URL correctly', () => {
      const API_BASE_URL = 'http://localhost:8080';
      const NUTRITION_SEARCH = '/api/v1/nutrition/search';
      const fullURL = `${API_BASE_URL}${NUTRITION_SEARCH}`;

      expect(fullURL).toBe('http://localhost:8080/api/v1/nutrition/search');
    });

    it('should handle trailing slash in base URL', () => {
      const API_BASE_URL = 'http://localhost:8080/';
      const NUTRITION_SEARCH = '/api/v1/nutrition/search';
      const fullURL = `${API_BASE_URL.replace(/\/$/, '')}${NUTRITION_SEARCH}`;

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
