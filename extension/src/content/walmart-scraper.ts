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
 *
 * Security: Uses string-based parsing instead of regex to prevent ReDoS attacks.
 * Input is also limited to 500 characters as an additional safeguard.
 */
function extractSizeFromText(text: string): string | undefined {
  // Limit input length to prevent DoS on extremely long strings
  const MAX_LENGTH = 500;
  const safeText = text.length > MAX_LENGTH ? text.slice(0, MAX_LENGTH) : text;
  const lowerText = safeText.toLowerCase();

  // Unit suffixes to look for (ordered by specificity - longer first)
  // Single letter units (l, g) need word boundary check
  const units: Array<{ unit: string; needsBoundary: boolean }> = [
    { unit: 'ounces', needsBoundary: false },
    { unit: 'ounce', needsBoundary: false },
    { unit: 'oz', needsBoundary: false },
    { unit: 'milliliters', needsBoundary: false },
    { unit: 'milliliter', needsBoundary: false },
    { unit: 'ml', needsBoundary: false },
    { unit: 'pounds', needsBoundary: false },
    { unit: 'pound', needsBoundary: false },
    { unit: 'lbs', needsBoundary: false },
    { unit: 'lb', needsBoundary: false },
    { unit: 'kilograms', needsBoundary: false },
    { unit: 'kilogram', needsBoundary: false },
    { unit: 'kg', needsBoundary: false },
    { unit: 'liters', needsBoundary: false },
    { unit: 'liter', needsBoundary: false },
    { unit: 'grams', needsBoundary: false },
    { unit: 'gram', needsBoundary: false },
    { unit: 'g', needsBoundary: true },  // Need boundary to avoid matching in words
    { unit: 'l', needsBoundary: true },  // Need boundary to avoid matching in words
  ];

  // Find each unit in the text and extract the number before it
  for (const { unit, needsBoundary } of units) {
    const unitIndex = lowerText.indexOf(unit);
    if (unitIndex > 0) {
      // Check word boundary after unit if needed
      if (needsBoundary) {
        const charAfter = lowerText[unitIndex + unit.length];
        if (charAfter && charAfter !== ' ' && charAfter !== ',' && charAfter !== ')' && charAfter !== '\n' && charAfter !== '\t') {
          continue; // Not a word boundary, skip
        }
      }

      // Look backwards from the unit to find the number
      const beforeUnit = safeText.slice(0, unitIndex);
      const numberMatch = extractNumberFromEnd(beforeUnit);
      if (numberMatch) {
        // Get the actual unit text with original casing
        const actualUnit = safeText.slice(unitIndex, unitIndex + unit.length);
        // Check if there was a space between number and unit
        const hasSpace = beforeUnit.endsWith(' ') || beforeUnit.endsWith('\t');
        return hasSpace ? `${numberMatch} ${actualUnit}` : `${numberMatch}${actualUnit}`;
      }
    }
  }

  // Check for "pack of N" format (preserve original case)
  const packOfIndex = lowerText.indexOf('pack of ');
  if (packOfIndex !== -1) {
    const afterPackOf = safeText.slice(packOfIndex + 8);
    const numberMatch = extractNumberFromStart(afterPackOf);
    if (numberMatch) {
      // Get "Pack of" with original casing
      const packOfText = safeText.slice(packOfIndex, packOfIndex + 8);
      return `${packOfText}${numberMatch}`;
    }
  }

  // Check for "N-pack" format
  const packIndex = lowerText.indexOf('-pack');
  if (packIndex > 0) {
    const beforePack = safeText.slice(0, packIndex);
    const numberMatch = extractNumberFromEnd(beforePack);
    if (numberMatch) {
      return `${numberMatch}-pack`;
    }
  }

  return undefined;
}

/**
 * Extracts a number (with optional decimal) from the end of a string.
 * Returns the number string or undefined if not found.
 */
function extractNumberFromEnd(text: string): string | undefined {
  let end = text.length;
  let start = end;

  // Skip trailing whitespace
  while (start > 0 && text[start - 1] === ' ') {
    start--;
    end = start;
  }

  // Find digits (and optional decimal point)
  let hasDecimal = false;
  while (start > 0) {
    const char = text[start - 1];
    if (char >= '0' && char <= '9') {
      start--;
    } else if (char === '.' && !hasDecimal) {
      hasDecimal = true;
      start--;
    } else {
      break;
    }
  }

  if (start < end) {
    const result = text.slice(start, end);
    // Make sure it starts with a digit (not just ".")
    if (result[0] >= '0' && result[0] <= '9') {
      return result;
    }
  }

  return undefined;
}

/**
 * Extracts a number from the start of a string.
 * Returns the number string or undefined if not found.
 */
function extractNumberFromStart(text: string): string | undefined {
  let index = 0;

  // Skip leading whitespace
  while (index < text.length && text[index] === ' ') {
    index++;
  }

  const start = index;
  let hasDecimal = false;

  // Find digits (and optional decimal point)
  while (index < text.length) {
    const char = text[index];
    if (char >= '0' && char <= '9') {
      index++;
    } else if (char === '.' && !hasDecimal) {
      hasDecimal = true;
      index++;
    } else {
      break;
    }
  }

  if (index > start) {
    const result = text.slice(start, index);
    // Make sure it starts with a digit
    if (result[0] >= '0' && result[0] <= '9') {
      return result;
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
