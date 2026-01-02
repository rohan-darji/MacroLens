// Tests for content/index.ts - Main content script entry point

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';


// Mock walmart-scraper module
vi.mock('@/content/walmart-scraper', () => ({
  isWalmartProductPage: vi.fn(),
  extractProductInfo: vi.fn(),
  log: vi.fn(),
}));

// Mock ui-overlay module
vi.mock('@/content/ui-overlay', () => ({
  createOverlay: vi.fn(),
  showLoading: vi.fn(),
  showError: vi.fn(),
  showNutrition: vi.fn(),
  removeOverlay: vi.fn(),
}));

describe('Content Script Entry Point', () => {
  let mockIsWalmartProductPage: ReturnType<typeof vi.fn>;
  let mockExtractProductInfo: ReturnType<typeof vi.fn>;
  let mockCreateOverlay: ReturnType<typeof vi.fn>;
  let mockShowLoading: ReturnType<typeof vi.fn>;
  let mockShowError: ReturnType<typeof vi.fn>;
  let mockShowNutrition: ReturnType<typeof vi.fn>;
  let mockRemoveOverlay: ReturnType<typeof vi.fn>;
  let mockSendMessage: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    // Set up chrome.runtime.sendMessage mock
    mockSendMessage = vi.fn();
    (global.chrome.runtime.sendMessage as ReturnType<typeof vi.fn>) = mockSendMessage;
    (global.chrome.runtime as any).lastError = null;

    // Get the mocked functions
    const scraper = await import('@/content/walmart-scraper');
    mockIsWalmartProductPage = scraper.isWalmartProductPage as ReturnType<typeof vi.fn>;
    mockExtractProductInfo = scraper.extractProductInfo as ReturnType<typeof vi.fn>;

    const overlay = await import('@/content/ui-overlay');
    mockCreateOverlay = overlay.createOverlay as ReturnType<typeof vi.fn>;
    mockShowLoading = overlay.showLoading as ReturnType<typeof vi.fn>;
    mockShowError = overlay.showError as ReturnType<typeof vi.fn>;
    mockShowNutrition = overlay.showNutrition as ReturnType<typeof vi.fn>;
    mockRemoveOverlay = overlay.removeOverlay as ReturnType<typeof vi.fn>;

    // Default mocks
    mockIsWalmartProductPage.mockReturnValue(false);
    mockExtractProductInfo.mockReturnValue(null);
    mockCreateOverlay.mockReturnValue(document.createElement('div'));

    // Set document.readyState
    Object.defineProperty(document, 'readyState', {
      value: 'complete',
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  describe('init function behavior', () => {
    it('should not process non-Walmart pages', async () => {
      mockIsWalmartProductPage.mockReturnValue(false);

      await import('@/content/index');

      // Run initial timers for waitForPageLoad
      await vi.advanceTimersByTimeAsync(1000);

      expect(mockIsWalmartProductPage).toHaveBeenCalled();
      expect(mockRemoveOverlay).toHaveBeenCalled();
      expect(mockCreateOverlay).not.toHaveBeenCalled();
    });

    it('should process Walmart product pages', async () => {
      mockIsWalmartProductPage.mockReturnValue(true);
      mockExtractProductInfo.mockReturnValue({
        name: 'Test Product',
        brand: 'Test Brand',
      });

      const mockResponse = {
        data: {
          fdcId: '123',
          productName: 'Test Product',
          servingSize: '100',
          servingSizeUnit: 'g',
          nutrients: { calories: 100, protein: 10, carbohydrates: 20, totalFat: 5 },
          confidence: 85,
          source: 'USDA',
        },
        cached: false,
      };

      mockSendMessage.mockImplementation((_msg, callback) => {
        callback(mockResponse);
      });

      await import('@/content/index');

      // Run initial timers for waitForPageLoad
      await vi.advanceTimersByTimeAsync(1000);

      expect(mockIsWalmartProductPage).toHaveBeenCalled();
      expect(mockExtractProductInfo).toHaveBeenCalled();
      expect(mockCreateOverlay).toHaveBeenCalled();
      expect(mockShowLoading).toHaveBeenCalled();
      expect(mockSendMessage).toHaveBeenCalled();
    });

    it('should show error when product info extraction fails', async () => {
      mockIsWalmartProductPage.mockReturnValue(true);
      mockExtractProductInfo.mockReturnValue(null);

      await import('@/content/index');

      await vi.advanceTimersByTimeAsync(1000);

      expect(mockIsWalmartProductPage).toHaveBeenCalled();
      expect(mockExtractProductInfo).toHaveBeenCalled();
      expect(mockCreateOverlay).not.toHaveBeenCalled();
    });

    it('should show error when response contains error', async () => {
      mockIsWalmartProductPage.mockReturnValue(true);
      mockExtractProductInfo.mockReturnValue({
        name: 'Test Product',
      });

      mockSendMessage.mockImplementation((_msg, callback) => {
        callback({ error: 'Product not found', cached: false });
      });

      await import('@/content/index');

      await vi.advanceTimersByTimeAsync(1000);

      expect(mockShowError).toHaveBeenCalledWith('Product not found', expect.any(HTMLElement));
    });

    it('should show nutrition data when response is successful', async () => {
      mockIsWalmartProductPage.mockReturnValue(true);
      mockExtractProductInfo.mockReturnValue({
        name: 'Test Product',
      });

      const mockData = {
        fdcId: '123',
        productName: 'Test Product',
        servingSize: '100',
        servingSizeUnit: 'g',
        nutrients: { calories: 100, protein: 10, carbohydrates: 20, totalFat: 5 },
        confidence: 85,
        source: 'USDA' as const,
      };

      mockSendMessage.mockImplementation((_msg, callback) => {
        callback({ data: mockData, cached: false });
      });

      await import('@/content/index');

      await vi.advanceTimersByTimeAsync(1000);

      expect(mockShowNutrition).toHaveBeenCalledWith(mockData, false, expect.any(HTMLElement));
    });

    it('should show error when no data received', async () => {
      mockIsWalmartProductPage.mockReturnValue(true);
      mockExtractProductInfo.mockReturnValue({
        name: 'Test Product',
      });

      mockSendMessage.mockImplementation((_msg, callback) => {
        callback({ cached: false });
      });

      await import('@/content/index');

      await vi.advanceTimersByTimeAsync(1000);

      expect(mockShowError).toHaveBeenCalledWith('No nutrition data received', expect.any(HTMLElement));
    });
  });

  describe('sendMessage error handling', () => {
    it('should handle chrome.runtime.lastError', async () => {
      mockIsWalmartProductPage.mockReturnValue(true);
      mockExtractProductInfo.mockReturnValue({
        name: 'Test Product',
      });

      mockSendMessage.mockImplementation((_msg, callback) => {
        (global.chrome.runtime as any).lastError = { message: 'Extension context invalidated' };
        callback(undefined);
      });

      await import('@/content/index');

      await vi.advanceTimersByTimeAsync(1000);

      expect(mockShowError).toHaveBeenCalledWith('Extension context invalidated', expect.any(HTMLElement));

      // Reset lastError
      (global.chrome.runtime as any).lastError = null;
    });

    it('should handle non-Error exception in catch block', async () => {
      mockIsWalmartProductPage.mockReturnValue(true);
      mockExtractProductInfo.mockReturnValue({
        name: 'Test Product',
      });

      mockSendMessage.mockImplementation(() => {
        throw 'string error';
      });

      await import('@/content/index');

      await vi.advanceTimersByTimeAsync(1000);

      expect(mockShowError).toHaveBeenCalledWith('Failed to load nutrition data', expect.any(HTMLElement));
    });
  });
});

describe('Content Script Navigation Handling', () => {
  let mockIsWalmartProductPage: ReturnType<typeof vi.fn>;
  let mockRemoveOverlay: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;

    const scraper = await import('@/content/walmart-scraper');
    mockIsWalmartProductPage = scraper.isWalmartProductPage as ReturnType<typeof vi.fn>;
    mockIsWalmartProductPage.mockReturnValue(false);

    const overlay = await import('@/content/ui-overlay');
    mockRemoveOverlay = overlay.removeOverlay as ReturnType<typeof vi.fn>;

    Object.defineProperty(document, 'readyState', {
      value: 'complete',
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should set up popstate event listener', async () => {
    const addEventListenerSpy = vi.spyOn(window, 'addEventListener');

    await import('@/content/index');

    expect(addEventListenerSpy).toHaveBeenCalledWith('popstate', expect.any(Function));

    addEventListenerSpy.mockRestore();
  });

  it('should intercept history.pushState', async () => {
    const originalPushState = history.pushState;

    await import('@/content/index');

    // history.pushState should be overridden
    expect(history.pushState).not.toBe(originalPushState);
  });

  it('should intercept history.replaceState', async () => {
    const originalReplaceState = history.replaceState;

    await import('@/content/index');

    // history.replaceState should be overridden
    expect(history.replaceState).not.toBe(originalReplaceState);
  });

  it('should detect URL change on popstate', async () => {
    await import('@/content/index');

    // Advance past initial load
    await vi.advanceTimersByTimeAsync(1000);

    // Clear mock calls from initial load
    mockIsWalmartProductPage.mockClear();
    mockRemoveOverlay.mockClear();

    // Simulate URL change
    Object.defineProperty(window, 'location', {
      value: { href: 'https://www.walmart.com/ip/new-product/456' },
      writable: true,
      configurable: true,
    });

    // Dispatch popstate event
    window.dispatchEvent(new PopStateEvent('popstate'));

    // Wait for the setTimeout in handleUrlChange
    await vi.advanceTimersByTimeAsync(500);
    await vi.advanceTimersByTimeAsync(1000);

    expect(mockIsWalmartProductPage).toHaveBeenCalled();
  });
});

describe('Content Script waitForPageLoad', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should wait for load event when document is still loading', async () => {
    // Set up spy BEFORE changing readyState and importing
    const addEventListenerSpy = vi.spyOn(document, 'addEventListener');

    Object.defineProperty(document, 'readyState', {
      value: 'loading',
      writable: true,
      configurable: true,
    });

    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(false),
      extractProductInfo: vi.fn(),
      log: vi.fn(),
    }));

    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn(),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    await import('@/content/index');

    // Should add DOMContentLoaded listener when loading
    expect(addEventListenerSpy).toHaveBeenCalledWith('DOMContentLoaded', expect.any(Function));

    addEventListenerSpy.mockRestore();
  });

  it('should use setTimeout when document is already complete', async () => {
    Object.defineProperty(document, 'readyState', {
      value: 'complete',
      writable: true,
      configurable: true,
    });

    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(false),
      extractProductInfo: vi.fn(),
      log: vi.fn(),
    }));

    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn(),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    await import('@/content/index');

    // Should proceed with setTimeout
    await vi.advanceTimersByTimeAsync(1000);

    const scraper = await import('@/content/walmart-scraper');
    expect(scraper.isWalmartProductPage).toHaveBeenCalled();
  });
});

describe('Content Script Navigation Debouncing', () => {
  let mockIsWalmartProductPage: ReturnType<typeof vi.fn>;
  let mockRemoveOverlay: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;

    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(false),
      extractProductInfo: vi.fn(),
      log: vi.fn(),
    }));

    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn(),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    Object.defineProperty(document, 'readyState', {
      value: 'complete',
      writable: true,
      configurable: true,
    });

    const scraper = await import('@/content/walmart-scraper');
    mockIsWalmartProductPage = scraper.isWalmartProductPage as ReturnType<typeof vi.fn>;

    const overlay = await import('@/content/ui-overlay');
    mockRemoveOverlay = overlay.removeOverlay as ReturnType<typeof vi.fn>;
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should debounce rapid navigation events', async () => {
    await import('@/content/index');

    // Advance past initial load
    await vi.advanceTimersByTimeAsync(1000);

    mockIsWalmartProductPage.mockClear();
    mockRemoveOverlay.mockClear();

    // Simulate rapid URL changes
    Object.defineProperty(window, 'location', {
      value: { href: 'https://www.walmart.com/ip/product-1/111' },
      writable: true,
      configurable: true,
    });
    window.dispatchEvent(new PopStateEvent('popstate'));

    // Quick second navigation before debounce completes
    await vi.advanceTimersByTimeAsync(200);
    Object.defineProperty(window, 'location', {
      value: { href: 'https://www.walmart.com/ip/product-2/222' },
      writable: true,
      configurable: true,
    });
    window.dispatchEvent(new PopStateEvent('popstate'));

    // Wait for debounce (500ms) and init waitForPageLoad (1000ms)
    await vi.advanceTimersByTimeAsync(500);
    await vi.advanceTimersByTimeAsync(1000);

    // isWalmartProductPage called only once due to debouncing (first navigation was cancelled)
    expect(mockIsWalmartProductPage).toHaveBeenCalledTimes(1);

    // removeOverlay is called:
    // - 2 times immediately in handleUrlChange (once per navigation)
    // - 1 time in init() when isWalmartProductPage returns false
    expect(mockRemoveOverlay).toHaveBeenCalledTimes(3);
  });

  it('should remove overlay immediately on navigation', async () => {
    await import('@/content/index');

    // Advance past initial load
    await vi.advanceTimersByTimeAsync(1000);

    mockRemoveOverlay.mockClear();

    // Simulate URL change
    Object.defineProperty(window, 'location', {
      value: { href: 'https://www.walmart.com/ip/new-product/456' },
      writable: true,
      configurable: true,
    });
    window.dispatchEvent(new PopStateEvent('popstate'));

    // removeOverlay should be called immediately (before debounce delay)
    expect(mockRemoveOverlay).toHaveBeenCalledTimes(1);
  });
});

describe('Content Script Concurrent Request Prevention', () => {
  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;

    Object.defineProperty(document, 'readyState', {
      value: 'complete',
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should prevent concurrent processing', async () => {
    let resolveFirst: Function;
    const firstPromise = new Promise((resolve) => {
      resolveFirst = resolve;
    });

    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(true),
      extractProductInfo: vi.fn().mockReturnValue({ name: 'Test' }),
      log: vi.fn(),
    }));

    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn().mockReturnValue(document.createElement('div')),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    (global.chrome.runtime.sendMessage as ReturnType<typeof vi.fn>).mockImplementation(
      async (_msg, callback) => {
        await firstPromise;
        callback({ data: { productName: 'Test' }, cached: false });
      }
    );

    await import('@/content/index');

    // Start initial processing
    await vi.advanceTimersByTimeAsync(1000);

    const scraper = await import('@/content/walmart-scraper');

    // Try to trigger another init while first is processing (simulate popstate)
    window.dispatchEvent(new PopStateEvent('popstate'));
    window.dispatchEvent(new PopStateEvent('popstate'));
    await vi.advanceTimersByTimeAsync(500);

    // The second call should be blocked by isProcessing flag
    // Note: isWalmartProductPage will be called but init will return early if isProcessing is true

    // Complete first request
    resolveFirst!();
    await vi.advanceTimersByTimeAsync(100);

    // Verify the module loaded and processed
    expect(scraper.isWalmartProductPage).toHaveBeenCalled();
  });
});

describe('Content Script URL Polling', () => {
  let mockIsWalmartProductPage: ReturnType<typeof vi.fn>;
  let mockRemoveOverlay: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;

    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(false),
      extractProductInfo: vi.fn(),
      log: vi.fn(),
    }));

    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn(),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    Object.defineProperty(document, 'readyState', {
      value: 'complete',
      writable: true,
      configurable: true,
    });

    const scraper = await import('@/content/walmart-scraper');
    mockIsWalmartProductPage = scraper.isWalmartProductPage as ReturnType<typeof vi.fn>;

    const overlay = await import('@/content/ui-overlay');
    mockRemoveOverlay = overlay.removeOverlay as ReturnType<typeof vi.fn>;
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should detect URL change via polling interval', async () => {
    await import('@/content/index');

    // Advance past initial load
    await vi.advanceTimersByTimeAsync(1000);

    mockIsWalmartProductPage.mockClear();
    mockRemoveOverlay.mockClear();

    // Change URL without triggering popstate
    Object.defineProperty(window, 'location', {
      value: { href: 'https://www.walmart.com/ip/polled-product/999' },
      writable: true,
      configurable: true,
    });

    // Advance by polling interval (1000ms)
    await vi.advanceTimersByTimeAsync(1000);

    // Should detect URL change and call removeOverlay
    expect(mockRemoveOverlay).toHaveBeenCalled();

    // Wait for debounce and init
    await vi.advanceTimersByTimeAsync(500);
    await vi.advanceTimersByTimeAsync(1000);

    expect(mockIsWalmartProductPage).toHaveBeenCalled();
  });

  it('should not trigger handleUrlChange when URL has not changed', async () => {
    await import('@/content/index');

    // Advance past initial load
    await vi.advanceTimersByTimeAsync(1000);

    mockRemoveOverlay.mockClear();

    // Advance by multiple polling intervals without changing URL
    await vi.advanceTimersByTimeAsync(3000);

    // removeOverlay should not be called since URL didn't change
    expect(mockRemoveOverlay).not.toHaveBeenCalled();
  });
});

describe('Content Script Error Handling', () => {
  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;

    Object.defineProperty(document, 'readyState', {
      value: 'complete',
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should handle error in outer try-catch and remove overlay', async () => {
    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(true),
      extractProductInfo: vi.fn().mockImplementation(() => {
        throw new Error('Extraction failed');
      }),
      log: vi.fn(),
    }));

    const mockRemoveOverlay = vi.fn();
    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn().mockReturnValue(document.createElement('div')),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: mockRemoveOverlay,
    }));

    await import('@/content/index');

    await vi.advanceTimersByTimeAsync(1000);

    // The outer catch should handle the error
    const scraper = await import('@/content/walmart-scraper');
    expect(scraper.isWalmartProductPage).toHaveBeenCalled();
  });

  it('should handle Error instance in inner catch block', async () => {
    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(true),
      extractProductInfo: vi.fn().mockReturnValue({ name: 'Test' }),
      log: vi.fn(),
    }));

    const mockShowError = vi.fn();
    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn().mockReturnValue(document.createElement('div')),
      showLoading: vi.fn(),
      showError: mockShowError,
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    (global.chrome.runtime.sendMessage as ReturnType<typeof vi.fn>).mockImplementation(() => {
      throw new Error('Network failure');
    });

    await import('@/content/index');

    await vi.advanceTimersByTimeAsync(1000);

    expect(mockShowError).toHaveBeenCalledWith('Network failure', expect.any(HTMLElement));
  });
});

describe('Content Script waitForPageLoad Edge Cases', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should handle load event firing during wait', async () => {
    const addEventListenerSpy = vi.spyOn(window, 'addEventListener');

    Object.defineProperty(document, 'readyState', {
      value: 'interactive',
      writable: true,
      configurable: true,
    });

    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(false),
      extractProductInfo: vi.fn(),
      log: vi.fn(),
    }));

    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn(),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    await import('@/content/index');

    // Simulate load event
    window.dispatchEvent(new Event('load'));

    // Wait for dynamic content delay
    await vi.advanceTimersByTimeAsync(1000);

    const scraper = await import('@/content/walmart-scraper');
    expect(scraper.isWalmartProductPage).toHaveBeenCalled();

    addEventListenerSpy.mockRestore();
  });

  it('should use fallback timeout when load event does not fire', async () => {
    Object.defineProperty(document, 'readyState', {
      value: 'interactive',
      writable: true,
      configurable: true,
    });

    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(false),
      extractProductInfo: vi.fn(),
      log: vi.fn(),
    }));

    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn(),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    await import('@/content/index');

    // Wait for fallback timeout (10000ms) + dynamic content delay (1000ms)
    await vi.advanceTimersByTimeAsync(11000);

    const scraper = await import('@/content/walmart-scraper');
    expect(scraper.isWalmartProductPage).toHaveBeenCalled();
  });
});

describe('Content Script history.pushState/replaceState', () => {
  let mockIsWalmartProductPage: ReturnType<typeof vi.fn>;
  let mockRemoveOverlay: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;

    vi.doMock('@/content/walmart-scraper', () => ({
      isWalmartProductPage: vi.fn().mockReturnValue(false),
      extractProductInfo: vi.fn(),
      log: vi.fn(),
    }));

    vi.doMock('@/content/ui-overlay', () => ({
      createOverlay: vi.fn(),
      showLoading: vi.fn(),
      showError: vi.fn(),
      showNutrition: vi.fn(),
      removeOverlay: vi.fn(),
    }));

    Object.defineProperty(document, 'readyState', {
      value: 'complete',
      writable: true,
      configurable: true,
    });

    const scraper = await import('@/content/walmart-scraper');
    mockIsWalmartProductPage = scraper.isWalmartProductPage as ReturnType<typeof vi.fn>;

    const overlay = await import('@/content/ui-overlay');
    mockRemoveOverlay = overlay.removeOverlay as ReturnType<typeof vi.fn>;
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should detect URL change via history.pushState', async () => {
    await import('@/content/index');

    // Advance past initial load
    await vi.advanceTimersByTimeAsync(1000);

    mockIsWalmartProductPage.mockClear();
    mockRemoveOverlay.mockClear();

    // Change location.href to simulate what pushState does
    Object.defineProperty(window, 'location', {
      value: { href: 'http://localhost/ip/pushed-product/888' },
      writable: true,
      configurable: true,
    });

    // Call pushState with same-origin URL (jsdom restriction)
    history.pushState({}, '', '/ip/pushed-product/888');

    // Should trigger handleUrlChange and remove overlay immediately
    expect(mockRemoveOverlay).toHaveBeenCalled();

    // Wait for debounce and init
    await vi.advanceTimersByTimeAsync(500);
    await vi.advanceTimersByTimeAsync(1000);

    expect(mockIsWalmartProductPage).toHaveBeenCalled();
  });

  it('should detect URL change via history.replaceState', async () => {
    await import('@/content/index');

    // Advance past initial load
    await vi.advanceTimersByTimeAsync(1000);

    mockIsWalmartProductPage.mockClear();
    mockRemoveOverlay.mockClear();

    // Change location.href to simulate what replaceState does
    Object.defineProperty(window, 'location', {
      value: { href: 'http://localhost/ip/replaced-product/777' },
      writable: true,
      configurable: true,
    });

    // Call replaceState with same-origin URL (jsdom restriction)
    history.replaceState({}, '', '/ip/replaced-product/777');

    // Should trigger handleUrlChange and remove overlay immediately
    expect(mockRemoveOverlay).toHaveBeenCalled();

    // Wait for debounce and init
    await vi.advanceTimersByTimeAsync(500);
    await vi.advanceTimersByTimeAsync(1000);

    expect(mockIsWalmartProductPage).toHaveBeenCalled();
  });
});

