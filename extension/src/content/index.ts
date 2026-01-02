// content/index.ts - Main content script entry point

import { isWalmartProductPage, extractProductInfo, log } from './walmart-scraper';
import { createOverlay, showLoading, showError, showNutrition, removeOverlay } from './ui-overlay';
import { MessageType, type GetNutritionPayload, type NutritionResponsePayload } from '@/types/messages';

// Track current URL to detect navigation
let currentUrl = window.location.href;
let isProcessing = false;

/**
 * Main initialization function
 */
async function init(): Promise<void> {
  // Prevent multiple simultaneous calls
  if (isProcessing) {
    return;
  }

  if (!isWalmartProductPage()) {
    log('Not a Walmart product page, extension inactive');
    removeOverlay(); // Clean up any existing overlay
    return;
  }

  log('Walmart product page detected!', window.location.href);
  isProcessing = true;

  let overlay;

  try {
    // Wait a bit for the page to fully load
    await waitForPageLoad();

    // Extract product information
    const productInfo = extractProductInfo();

    if (!productInfo) {
      log('Failed to extract product information');
      return;
    }

    log('Product info extracted:', productInfo);

    // Create overlay with loading state
    overlay = createOverlay();
    showLoading(overlay);

    // Send message to background script to fetch nutrition data
    try {
      const payload: GetNutritionPayload = {
        productInfo,
        timestamp: Date.now(),
      };

      const response = await sendMessage<NutritionResponsePayload>({
        type: MessageType.GET_NUTRITION,
        payload,
      });

      if (response.error) {
        showError(response.error, overlay);
      } else if (response.data) {
        showNutrition(response.data, response.cached, overlay);
      } else {
        showError('No nutrition data received', overlay);
      }
    } catch (error) {
      log('Error fetching nutrition data:', error);
      showError(
        error instanceof Error ? error.message : 'Failed to load nutrition data',
        overlay
      );
    }
  } catch (error) {
    log('Error during initialization:', error);
    if (overlay) {
      removeOverlay();
    }
  } finally {
    isProcessing = false;
  }
}

/**
 * Handle URL changes (SPA navigation)
 */
function handleUrlChange(): void {
  const newUrl = window.location.href;
  if (newUrl !== currentUrl) {
    log('URL changed:', currentUrl, '->', newUrl);
    currentUrl = newUrl;
    // Small delay to let the page content update
    setTimeout(() => init(), 500);
  }
}

/**
 * Set up listeners for SPA navigation
 */
function setupNavigationListeners(): void {
  // Listen for browser back/forward navigation
  window.addEventListener('popstate', handleUrlChange);

  // Override history.pushState to detect programmatic navigation
  const originalPushState = history.pushState;
  history.pushState = function (...args) {
    originalPushState.apply(this, args);
    handleUrlChange();
  };

  // Override history.replaceState
  const originalReplaceState = history.replaceState;
  history.replaceState = function (...args) {
    originalReplaceState.apply(this, args);
    handleUrlChange();
  };

  // Also use MutationObserver as fallback for dynamic content changes
  const observer = new MutationObserver(() => {
    // Check if URL changed (some SPAs update URL without using history API)
    if (window.location.href !== currentUrl) {
      handleUrlChange();
    }
  });

  // Observe changes to the document
  observer.observe(document.body, {
    childList: true,
    subtree: true,
  });

  log('Navigation listeners set up');
}

const DEFAULT_DYNAMIC_CONTENT_DELAY_MS = 1000;
const DEFAULT_MAX_PAGE_LOAD_WAIT_MS = 10000;

type WaitForPageLoadOptions = {
  /**
   * Additional delay after the page is reported as loaded, to allow
   * dynamic content to render (in milliseconds).
   */
  dynamicContentDelayMs?: number;
  /**
   * Maximum time to wait for the page load event before proceeding
   * anyway (in milliseconds).
   */
  maxWaitMs?: number;
};

/**
 * Waits for the page to be fully loaded and ready for scraping.
 *
 * Defaults preserve the previous behavior (1s post-load delay) but
 * allow callers/configuration to override the timing if needed.
 */
function waitForPageLoad(options: WaitForPageLoadOptions = {}): Promise<void> {
  const {
    dynamicContentDelayMs = DEFAULT_DYNAMIC_CONTENT_DELAY_MS,
    maxWaitMs = DEFAULT_MAX_PAGE_LOAD_WAIT_MS,
  } = options;

  return new Promise((resolve) => {
    let resolved = false;
    let fallbackTimeoutId: number | null = null;

    const resolveOnce = () => {
      if (resolved) {
        return;
      }
      resolved = true;
      window.removeEventListener('load', onLoad);
      if (fallbackTimeoutId !== null) {
        clearTimeout(fallbackTimeoutId);
      }
      // Small delay to allow dynamic content to finish rendering
      setTimeout(resolve, dynamicContentDelayMs);
    };

    const onLoad = () => {
      resolveOnce();
    };

    if (document.readyState === 'complete') {
      // Page already loaded, just apply the dynamic content delay
      resolveOnce();
    } else {
      // Wait for the load event, but don't wait forever
      window.addEventListener('load', onLoad, { once: true });
      fallbackTimeoutId = window.setTimeout(() => {
        resolveOnce();
      }, maxWaitMs);
    }
  });
}

/**
 * Sends a message to the background script and waits for response
 */
function sendMessage<T>(message: any): Promise<T> {
  return new Promise((resolve, reject) => {
    chrome.runtime.sendMessage(message, (response) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message));
      } else {
        resolve(response);
      }
    });
  });
}

// Run initialization when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    setupNavigationListeners();
    init();
  });
} else {
  setupNavigationListeners();
  init();
}

log('Content script loaded');
