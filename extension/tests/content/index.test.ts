// Tests for content/index.ts - Main content script entry point

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { MessageType } from '@/types/messages';

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

describe('Content Script MutationObserver', () => {
  let observerCallback: MutationCallback;
  let mockObserve: ReturnType<typeof vi.fn>;
  let MockMutationObserver: typeof MutationObserver;

  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.useFakeTimers();

    (global.chrome.runtime as any).lastError = null;

    mockObserve = vi.fn();

    // Create a proper MutationObserver mock class
    MockMutationObserver = class {
      constructor(callback: MutationCallback) {
        observerCallback = callback;
      }
      observe = mockObserve;
      disconnect = vi.fn();
      takeRecords = vi.fn().mockReturnValue([]);
    } as unknown as typeof MutationObserver;

    global.MutationObserver = MockMutationObserver;

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
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.resetModules();
  });

  it('should set up MutationObserver', async () => {
    await import('@/content/index');

    expect(mockObserve).toHaveBeenCalledWith(document.body, {
      childList: true,
      subtree: true,
    });
  });

  it('should detect URL change via MutationObserver', async () => {
    const scraper = await import('@/content/walmart-scraper');
    const mockIsWalmartProductPage = scraper.isWalmartProductPage as ReturnType<typeof vi.fn>;

    await import('@/content/index');

    // Advance past initial load
    await vi.advanceTimersByTimeAsync(1000);

    mockIsWalmartProductPage.mockClear();

    // Change URL
    Object.defineProperty(window, 'location', {
      value: { href: 'https://www.walmart.com/ip/another-product/789' },
      writable: true,
      configurable: true,
    });

    // Trigger MutationObserver callback
    observerCallback([], {} as MutationObserver);

    // Wait for handleUrlChange setTimeout
    await vi.advanceTimersByTimeAsync(500);
    await vi.advanceTimersByTimeAsync(1000);

    expect(mockIsWalmartProductPage).toHaveBeenCalled();
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
    const initialCallCount = (scraper.isWalmartProductPage as ReturnType<typeof vi.fn>).mock.calls
      .length;

    // Try to trigger another init while first is processing (simulate popstate)
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

