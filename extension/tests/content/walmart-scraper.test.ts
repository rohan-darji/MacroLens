import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock the constants
vi.mock('@/config/constants', () => ({
  WALMART: {
    PRODUCT_URL_PATTERN: /^https:\/\/www\.walmart\.com\/ip\/[^\/?#]+\/\d+(?:[?#]|$)/,
  },
  DEBUG_MODE: false,
}));

// Test the extraction functions with DOM mocking
describe('Walmart Scraper', () => {
  let originalWindow: typeof window;
  let originalDocument: typeof document;

  beforeEach(() => {
    // Save originals
    originalWindow = global.window;
    originalDocument = global.document;
  });

  afterEach(() => {
    // Restore originals
    global.window = originalWindow;
    global.document = originalDocument;
    vi.clearAllMocks();
  });

  describe('isWalmartProductPage', () => {
    const createMockWindow = (href: string) => ({
      location: { href },
    });

    it('should return true for valid Walmart product URL', async () => {
      global.window = createMockWindow(
        'https://www.walmart.com/ip/Great-Value-Whole-Milk/10450114'
      ) as any;

      const { isWalmartProductPage } = await import('@/content/walmart-scraper');
      expect(isWalmartProductPage()).toBe(true);
    });

    it('should return false for Walmart homepage', async () => {
      global.window = createMockWindow('https://www.walmart.com/') as any;

      const { isWalmartProductPage } = await import('@/content/walmart-scraper');
      expect(isWalmartProductPage()).toBe(false);
    });

    it('should return false for Walmart search page', async () => {
      global.window = createMockWindow('https://www.walmart.com/search?q=milk') as any;

      const { isWalmartProductPage } = await import('@/content/walmart-scraper');
      expect(isWalmartProductPage()).toBe(false);
    });

    it('should return true for product URL with query params', async () => {
      global.window = createMockWindow(
        'https://www.walmart.com/ip/Product-Name/12345?ref=homepage'
      ) as any;

      const { isWalmartProductPage } = await import('@/content/walmart-scraper');
      expect(isWalmartProductPage()).toBe(true);
    });

    it('should return true for product URL with hash', async () => {
      global.window = createMockWindow(
        'https://www.walmart.com/ip/Product-Name/12345#reviews'
      ) as any;

      const { isWalmartProductPage } = await import('@/content/walmart-scraper');
      expect(isWalmartProductPage()).toBe(true);
    });
  });

  describe('extractProductInfo', () => {
    const createMockDocument = (jsonLdData: any = null, domElements: Record<string, string> = {}) => {
      const scripts: HTMLScriptElement[] = [];

      if (jsonLdData) {
        scripts.push({
          textContent: JSON.stringify(jsonLdData),
        } as HTMLScriptElement);
      }

      return {
        querySelectorAll: vi.fn((selector: string) => {
          if (selector === 'script[type="application/ld+json"]') {
            return scripts;
          }
          return [];
        }),
        querySelector: vi.fn((selector: string) => {
          if (domElements[selector]) {
            return {
              textContent: domElements[selector],
            };
          }
          return null;
        }),
        createElement: vi.fn(() => ({
          textContent: '',
          innerHTML: '',
          id: '',
          className: '',
          appendChild: vi.fn(),
          remove: vi.fn(),
        })),
        getElementById: vi.fn(() => null),
        body: {
          appendChild: vi.fn(),
        },
      };
    };

    it('should extract product info from JSON-LD data', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Great-Value-Milk/12345' },
      } as any;

      global.document = createMockDocument({
        '@type': 'Product',
        name: 'Great Value Whole Milk, Gallon',
        brand: { name: 'Great Value' },
        gtin13: '0078742001234',
      }) as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result).not.toBeNull();
      expect(result?.name).toBe('Great Value Whole Milk, Gallon');
      expect(result?.brand).toBe('Great Value');
      expect(result?.upc).toBe('0078742001234');
      expect(result?.retailer).toBe('walmart');
    });

    it('should extract product info with string brand', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createMockDocument({
        '@type': 'Product',
        name: 'Test Product',
        brand: 'Test Brand',
      }) as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result).not.toBeNull();
      expect(result?.brand).toBe('Test Brand');
    });

    it('should extract from @graph structure', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createMockDocument({
        '@graph': [
          { '@type': 'WebPage' },
          { '@type': 'Product', name: 'Nested Product' },
        ],
      }) as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result).not.toBeNull();
      expect(result?.name).toBe('Nested Product');
    });

    it('should fallback to DOM extraction when JSON-LD is missing', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createMockDocument(null, {
        'h1[itemprop="name"]': 'DOM Product Name',
        '[itemprop="brand"]': 'DOM Brand',
      }) as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result).not.toBeNull();
      expect(result?.name).toBe('DOM Product Name');
      expect(result?.brand).toBe('DOM Brand');
    });

    it('should return null for non-product pages', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/search?q=milk' },
      } as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result).toBeNull();
    });

    it('should return null when no product info found', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createMockDocument(null, {}) as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result).toBeNull();
    });

    it('should handle JSON-LD arrays', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createMockDocument([
        { '@type': 'BreadcrumbList' },
        { '@type': 'Product', name: 'Array Product' },
      ]) as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result).not.toBeNull();
      expect(result?.name).toBe('Array Product');
    });

    it('should use SKU when gtin13 is not available', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createMockDocument({
        '@type': 'Product',
        name: 'Test Product',
        sku: 'SKU123456',
      }) as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result).not.toBeNull();
      expect(result?.upc).toBe('SKU123456');
    });
  });

  describe('Size extraction from text', () => {
    const createDomOnlyDocument = (name: string) => ({
      querySelectorAll: vi.fn(() => []),
      querySelector: vi.fn((selector: string) => {
        if (selector === 'h1') {
          return { textContent: name };
        }
        return null;
      }),
      createElement: vi.fn(() => ({
        textContent: '',
        innerHTML: '',
      })),
      getElementById: vi.fn(() => null),
      body: { appendChild: vi.fn() },
    });

    it('should extract ounces from product name', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createDomOnlyDocument('Great Value Milk 128 oz') as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result?.size).toBe('128 oz');
    });

    it('should extract milliliters from product name', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createDomOnlyDocument('Coca-Cola 500ml') as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result?.size).toBe('500ml');
    });

    it('should extract pounds from product name', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      // Use "lbs" to avoid matching the liters pattern "l" first
      global.document = createDomOnlyDocument('Ground Beef 2 lbs') as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result?.size).toBe('2 lbs');
    });

    it('should extract grams from product name', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createDomOnlyDocument('Chips 200g bag') as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result?.size).toBe('200g');
    });

    it('should extract pack of format', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createDomOnlyDocument('Water Bottles Pack of 24') as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result?.size).toBe('Pack of 24');
    });

    it('should extract pack format with dash', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createDomOnlyDocument('Soda 12-pack') as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result?.size).toBe('12-pack');
    });

    it('should handle decimal sizes', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      global.document = createDomOnlyDocument('Cheese 1.5 lbs block') as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      const result = extractProductInfo();

      expect(result?.size).toBe('1.5 lbs');
    });

    it('should safely handle very long input strings (ReDoS prevention)', async () => {
      global.window = {
        location: { href: 'https://www.walmart.com/ip/Product/12345' },
      } as any;

      // Create a very long string that could cause ReDoS with vulnerable regex
      const longString = 'Product Name ' + '1'.repeat(1000) + ' oz';
      global.document = createDomOnlyDocument(longString) as any;

      const { extractProductInfo } = await import('@/content/walmart-scraper');
      
      // Should complete quickly without hanging (input is truncated to 500 chars)
      const startTime = Date.now();
      const result = extractProductInfo();
      const elapsed = Date.now() - startTime;

      // Should complete in under 100ms (ReDoS would take much longer)
      expect(elapsed).toBeLessThan(100);
      // Won't find size because it's truncated
      expect(result?.size).toBeUndefined();
    });
  });

  describe('log function', () => {
    it('should not throw when DEBUG_MODE is false', async () => {
      const { log } = await import('@/content/walmart-scraper');
      expect(() => log('test message')).not.toThrow();
    });
  });
});
