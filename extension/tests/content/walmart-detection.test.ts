import { describe, it, expect } from 'vitest';

// Mock constants
const WALMART = {
  PRODUCT_URL_PATTERN: /^https:\/\/www\.walmart\.com\/ip\/[^\/?#]+\/\d+(?:[?#]|$)/,
  PRODUCT_PAGE_INDICATOR: '/ip/',
};

// Function under test (copied from content/index.ts)
function isWalmartProductPage(url: string): boolean {
  return WALMART.PRODUCT_URL_PATTERN.test(url);
}

describe('Walmart Product Page Detection', () => {
  describe('isWalmartProductPage', () => {
    it('should detect valid Walmart product page with full pattern', () => {
      const url = 'https://www.walmart.com/ip/Great-Value-Whole-Milk-Gallon/10450114';
      expect(isWalmartProductPage(url)).toBe(true);
    });

    it('should detect product page with different product name', () => {
      const url = 'https://www.walmart.com/ip/Some-Product-Name/12345678';
      expect(isWalmartProductPage(url)).toBe(true);
    });

    it('should detect product page with dashes in name', () => {
      const url = 'https://www.walmart.com/ip/Product-With-Many-Dashes/99999999';
      expect(isWalmartProductPage(url)).toBe(true);
    });

    it('should reject Walmart homepage', () => {
      const url = 'https://www.walmart.com/';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should reject Walmart search page', () => {
      const url = 'https://www.walmart.com/search?q=milk';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should reject Walmart category page', () => {
      const url = 'https://www.walmart.com/browse/food/dairy';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should reject non-Walmart URL', () => {
      const url = 'https://www.amazon.com/product/123';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should reject malformed URL', () => {
      const url = 'https://www.walmart.com/ip/';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should reject URL without product ID', () => {
      const url = 'https://www.walmart.com/ip/Product-Name';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should handle URL with query parameters', () => {
      const url = 'https://www.walmart.com/ip/Product-Name/12345?param=value';
      expect(isWalmartProductPage(url)).toBe(true);
    });

    it('should handle URL with hash', () => {
      const url = 'https://www.walmart.com/ip/Product-Name/12345#reviews';
      expect(isWalmartProductPage(url)).toBe(true);
    });

    it('should be case sensitive for protocol', () => {
      const url = 'HTTP://www.walmart.com/ip/Product-Name/12345';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should reject http (non-secure)', () => {
      const url = 'http://www.walmart.com/ip/Product-Name/12345';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should reject URL with extra path after product ID', () => {
      const url = 'https://www.walmart.com/ip/Product-Name/12345/extra-path';
      expect(isWalmartProductPage(url)).toBe(false);
    });

    it('should reject URL with slash in product name', () => {
      const url = 'https://www.walmart.com/ip/Product/Name/12345';
      expect(isWalmartProductPage(url)).toBe(false);
    });
  });
});
