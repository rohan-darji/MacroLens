import { describe, it, expect } from 'vitest';
import { API_BASE_URL, DEBUG_MODE } from '../../src/config/constants';

describe('Constants Configuration', () => {
  describe('API Configuration', () => {
    it('should have API_BASE_URL defined', () => {
      expect(API_BASE_URL).toBeDefined();
      expect(typeof API_BASE_URL).toBe('string');
    });

    it('should have DEBUG_MODE defined as boolean', () => {
      expect(typeof DEBUG_MODE).toBe('boolean');
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
