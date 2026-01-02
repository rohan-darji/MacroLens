// walmart-scraper.ts - Extract product information from Walmart pages

import type { ProductInfo } from '@/types/product';
import { WALMART, DEBUG_MODE } from '@/config/constants';

/**
 * JSON-LD structured data interface for Walmart products
 */
interface JsonLdProduct {
  '@type': string;
  name?: string;
  brand?: {
    name?: string;
  } | string;
  description?: string;
  gtin13?: string;
  sku?: string;
}

/**
 * Checks if the current page is a Walmart product page
 */
export function isWalmartProductPage(): boolean {
  const url = window.location.href;
  return WALMART.PRODUCT_URL_PATTERN.test(url);
}

/**
 * Extracts product information from the Walmart product page
 * Uses JSON-LD structured data as primary method, with DOM fallback
 */
export function extractProductInfo(): ProductInfo | null {
  if (!isWalmartProductPage()) {
    return null;
  }

  // Try JSON-LD structured data first (most reliable)
  const jsonLdProduct = extractFromJsonLd();
  if (jsonLdProduct) {
    return jsonLdProduct;
  }

  // Fallback to DOM extraction
  const domProduct = extractFromDom();
  if (domProduct) {
    return domProduct;
  }

  return null;
}

/**
 * Extracts product information from JSON-LD structured data
 * Walmart embeds product data in <script type="application/ld+json">
 */
function extractFromJsonLd(): ProductInfo | null {
  try {
    const jsonLdScripts = document.querySelectorAll('script[type="application/ld+json"]');

    for (const script of Array.from(jsonLdScripts)) {
      try {
        const data = JSON.parse(script.textContent || '');

        // Handle both single objects and arrays
        const products = Array.isArray(data) ? data : [data];

        for (const item of products) {
          // Look for Product type
          if (item['@type'] === 'Product' || item['@type']?.includes('Product')) {
            return parseJsonLdProduct(item);
          }

          // Check nested graph structure
          if (item['@graph']) {
            for (const graphItem of item['@graph']) {
              if (graphItem['@type'] === 'Product') {
                return parseJsonLdProduct(graphItem);
              }
            }
          }
        }
      } catch (parseError) {
        // Continue to next script if this one fails
        if (DEBUG_MODE) {
          console.warn('[MacroLens] Failed to parse JSON-LD:', parseError);
        }
      }
    }
  } catch (error) {
    if (DEBUG_MODE) {
      console.error('[MacroLens] Error extracting JSON-LD:', error);
    }
  }

  return null;
}

/**
 * Parses a JSON-LD product object into ProductInfo
 */
function parseJsonLdProduct(product: JsonLdProduct): ProductInfo | null {
  const name = product.name;
  if (!name) {
    return null;
  }

  // Extract brand (can be object or string)
  let brand: string | undefined;
  if (product.brand) {
    if (typeof product.brand === 'string') {
      brand = product.brand;
    } else if (product.brand.name) {
      brand = product.brand.name;
    }
  }

  return {
    name: name.trim(),
    brand: brand?.trim(),
    upc: product.gtin13 || product.sku,
    url: window.location.href,
    retailer: 'walmart',
  };
}

/**
 * Fallback: Extract product information from DOM selectors
 * Used when JSON-LD is not available or incomplete
 */
function extractFromDom(): ProductInfo | null {
  try {
    // Try multiple selectors for product name (Walmart's layout may vary)
    const nameSelectors = [
      'h1[itemprop="name"]',
      'h1[data-automation-id="product-title"]',
      'h1.prod-ProductTitle',
      'h1.f1',
      '[data-testid="product-title"]',
      'h1',
    ];

    let name: string | null = null;
    for (const selector of nameSelectors) {
      const element = document.querySelector(selector);
      if (element?.textContent?.trim()) {
        name = element.textContent.trim();
        break;
      }
    }

    if (!name) {
      return null;
    }

    // Try to extract brand
    const brandSelectors = [
      '[itemprop="brand"]',
      '[data-automation-id="product-brand"]',
      '.prod-BrandName',
      'a[link-identifier="Brand name"]',
    ];

    let brand: string | undefined;
    for (const selector of brandSelectors) {
      const element = document.querySelector(selector);
      if (element?.textContent?.trim()) {
        brand = element.textContent.trim();
        break;
      }
    }

    // Try to extract size/quantity from name or description
    const size = extractSizeFromText(name);

    return {
      name,
      brand,
      size,
      url: window.location.href,
      retailer: 'walmart',
    };
  } catch (error) {
    if (DEBUG_MODE) {
      console.error('[MacroLens] Error extracting from DOM:', error);
    }
    return null;
  }
}

/**
 * Attempts to extract size/quantity information from product text
 * Examples: "12 oz", "500ml", "1 lb", "pack of 6"
 */
function extractSizeFromText(text: string): string | undefined {
  // Order matters: more specific patterns must come before less specific ones
  // Within alternations, longer strings must come first (lbs before lb)
  const sizePatterns = [
    /(\d+\.?\d*\s*(?:ounces|ounce|oz))/i,
    /(\d+\.?\d*\s*(?:milliliters|milliliter|ml))/i,
    /(\d+\.?\d*\s*(?:pounds|pound|lbs|lb))/i,  // Must be before 'l' pattern
    /(\d+\.?\d*\s*(?:kilograms|kilogram|kg))/i, // Must be before 'g' pattern
    /(\d+\.?\d*\s*(?:liters|liter|l)(?:\s|$))/i, // Word boundary for standalone 'l'
    /(\d+\.?\d*\s*(?:grams|gram|g)(?:\s|$))/i,   // Word boundary for standalone 'g'
    /(pack of \d+)/i,
    /(\d+-pack)/i,
  ];

  for (const pattern of sizePatterns) {
    const match = text.match(pattern);
    if (match) {
      return match[1].trim();
    }
  }

  return undefined;
}

/**
 * Logs debug messages if DEBUG_MODE is enabled
 */
export function log(...args: any[]): void {
  if (DEBUG_MODE) {
    console.log('[MacroLens Scraper]', ...args);
  }
}
