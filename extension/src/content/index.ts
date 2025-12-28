// content/index.ts - Main content script entry point

import { WALMART, DEBUG_MODE } from '@/config/constants';

/**
 * Checks if the current page is a Walmart product page
 */
function isWalmartProductPage(): boolean {
  const url = window.location.href;
  return WALMART.PRODUCT_URL_PATTERN.test(url);
}

/**
 * Logs debug messages if DEBUG_MODE is enabled
 */
function log(...args: any[]): void {
  if (DEBUG_MODE) {
    console.log('[MacroLens]', ...args);
  }
}

/**
 * Main initialization function
 */
function init(): void {
  if (isWalmartProductPage()) {
    log('Walmart product page detected!', window.location.href);
    log('MacroLens extension is active');

    // TODO: In Phase 2, we will:
    // 1. Extract product information
    // 2. Send message to background script
    // 3. Display nutrition overlay
  } else {
    log('Not a product page, extension inactive');
  }
}

// Run initialization when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}

log('Content script loaded');
