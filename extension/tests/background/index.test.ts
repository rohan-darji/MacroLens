// Tests for background/index.ts - Background service worker

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { MessageType } from '@/types/messages';

// Mock the api-client module
vi.mock('@/background/api-client', () => ({
  searchNutrition: vi.fn(),
  clearCache: vi.fn(),
  getCacheStats: vi.fn(),
}));

// Mock the constants module
vi.mock('@/config/constants', () => ({
  DEBUG_MODE: false,
}));

describe('Background Service Worker', () => {
  let messageListener: Function;
  let mockSearchNutrition: ReturnType<typeof vi.fn>;
  let mockClearCache: ReturnType<typeof vi.fn>;
  let mockGetCacheStats: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();

    // Set up chrome mock to capture the message listener
    (global.chrome.runtime.onMessage.addListener as ReturnType<typeof vi.fn>).mockImplementation(
      (listener: Function) => {
        messageListener = listener;
      }
    );

    // Get the mocked functions
    const apiClient = await import('@/background/api-client');
    mockSearchNutrition = apiClient.searchNutrition as ReturnType<typeof vi.fn>;
    mockClearCache = apiClient.clearCache as ReturnType<typeof vi.fn>;
    mockGetCacheStats = apiClient.getCacheStats as ReturnType<typeof vi.fn>;

    // Import the module to register the listener
    await import('@/background/index');
  });

  afterEach(() => {
    vi.resetModules();
  });

  describe('Message Listener Registration', () => {
    it('should register a message listener on load', () => {
      expect(chrome.runtime.onMessage.addListener).toHaveBeenCalled();
      expect(messageListener).toBeDefined();
    });
  });

  describe('GET_NUTRITION Message Handling', () => {
    it('should handle GET_NUTRITION message and return nutrition data', async () => {
      const mockNutritionData = {
        fdcId: '123',
        productName: 'Test Product',
        servingSize: '100',
        servingSizeUnit: 'g',
        nutrients: { calories: 100, protein: 10, carbohydrates: 20, totalFat: 5 },
        confidence: 85,
        source: 'USDA' as const,
      };

      mockSearchNutrition.mockResolvedValue({
        data: mockNutritionData,
        cached: false,
      });

      const sendResponse = vi.fn();
      const message = {
        type: MessageType.GET_NUTRITION,
        payload: {
          productInfo: { name: 'Test Product', brand: 'Test Brand' },
          timestamp: Date.now(),
        },
      };

      const result = messageListener(message, {}, sendResponse);

      // Should return true to keep the message channel open
      expect(result).toBe(true);

      // Wait for async operation
      await vi.waitFor(() => {
        expect(sendResponse).toHaveBeenCalled();
      });

      expect(mockSearchNutrition).toHaveBeenCalledWith({
        name: 'Test Product',
        brand: 'Test Brand',
      });
      expect(sendResponse).toHaveBeenCalledWith({
        data: mockNutritionData,
        error: undefined,
        cached: false,
      });
    });

    it('should handle GET_NUTRITION with error from searchNutrition', async () => {
      mockSearchNutrition.mockResolvedValue({
        error: 'Product not found',
        cached: false,
      });

      const sendResponse = vi.fn();
      const message = {
        type: MessageType.GET_NUTRITION,
        payload: {
          productInfo: { name: 'Unknown Product' },
          timestamp: Date.now(),
        },
      };

      messageListener(message, {}, sendResponse);

      await vi.waitFor(() => {
        expect(sendResponse).toHaveBeenCalled();
      });

      expect(sendResponse).toHaveBeenCalledWith({
        data: undefined,
        error: 'Product not found',
        cached: false,
      });
    });

    it('should handle GET_NUTRITION with exception', async () => {
      mockSearchNutrition.mockRejectedValue(new Error('Network error'));

      const sendResponse = vi.fn();
      const message = {
        type: MessageType.GET_NUTRITION,
        payload: {
          productInfo: { name: 'Test Product' },
          timestamp: Date.now(),
        },
      };

      messageListener(message, {}, sendResponse);

      await vi.waitFor(() => {
        expect(sendResponse).toHaveBeenCalled();
      });

      expect(sendResponse).toHaveBeenCalledWith({
        error: 'Network error',
        cached: false,
      });
    });

    it('should handle non-Error exception in GET_NUTRITION', async () => {
      mockSearchNutrition.mockRejectedValue('string error');

      const sendResponse = vi.fn();
      const message = {
        type: MessageType.GET_NUTRITION,
        payload: {
          productInfo: { name: 'Test Product' },
          timestamp: Date.now(),
        },
      };

      messageListener(message, {}, sendResponse);

      await vi.waitFor(() => {
        expect(sendResponse).toHaveBeenCalled();
      });

      expect(sendResponse).toHaveBeenCalledWith({
        error: 'Unknown error occurred',
        cached: false,
      });
    });

    it('should handle cached nutrition data', async () => {
      const mockNutritionData = {
        fdcId: '456',
        productName: 'Cached Product',
        servingSize: '50',
        servingSizeUnit: 'g',
        nutrients: { calories: 50, protein: 5, carbohydrates: 10, totalFat: 2 },
        confidence: 90,
        source: 'Cache' as const,
      };

      mockSearchNutrition.mockResolvedValue({
        data: mockNutritionData,
        cached: true,
      });

      const sendResponse = vi.fn();
      const message = {
        type: MessageType.GET_NUTRITION,
        payload: {
          productInfo: { name: 'Cached Product' },
          timestamp: Date.now(),
        },
      };

      messageListener(message, {}, sendResponse);

      await vi.waitFor(() => {
        expect(sendResponse).toHaveBeenCalled();
      });

      expect(sendResponse).toHaveBeenCalledWith({
        data: mockNutritionData,
        error: undefined,
        cached: true,
      });
    });
  });

  describe('CLEAR_CACHE Message Handling', () => {
    it('should handle CLEAR_CACHE message successfully', async () => {
      mockClearCache.mockResolvedValue(undefined);
      mockGetCacheStats.mockResolvedValue({ count: 0, keys: [] });

      const sendResponse = vi.fn();
      const message = { type: MessageType.CLEAR_CACHE };

      const result = messageListener(message, {}, sendResponse);

      expect(result).toBe(true);

      await vi.waitFor(() => {
        expect(sendResponse).toHaveBeenCalled();
      });

      expect(mockClearCache).toHaveBeenCalled();
      expect(mockGetCacheStats).toHaveBeenCalled();
      expect(sendResponse).toHaveBeenCalledWith({
        success: true,
        message: 'Cache cleared successfully',
        stats: { count: 0, keys: [] },
      });
    });

    it('should handle CLEAR_CACHE with error', async () => {
      mockClearCache.mockRejectedValue(new Error('Storage error'));

      const sendResponse = vi.fn();
      const message = { type: MessageType.CLEAR_CACHE };

      messageListener(message, {}, sendResponse);

      await vi.waitFor(() => {
        expect(sendResponse).toHaveBeenCalled();
      });

      expect(sendResponse).toHaveBeenCalledWith({
        success: false,
        error: 'Storage error',
      });
    });

    it('should handle CLEAR_CACHE with non-Error exception', async () => {
      mockClearCache.mockRejectedValue('string error');

      const sendResponse = vi.fn();
      const message = { type: MessageType.CLEAR_CACHE };

      messageListener(message, {}, sendResponse);

      await vi.waitFor(() => {
        expect(sendResponse).toHaveBeenCalled();
      });

      expect(sendResponse).toHaveBeenCalledWith({
        success: false,
        error: 'Failed to clear cache',
      });
    });
  });

  describe('Unknown Message Type', () => {
    it('should handle unknown message type', () => {
      const sendResponse = vi.fn();
      const message = { type: 'UNKNOWN_TYPE' };

      const result = messageListener(message, {}, sendResponse);

      expect(result).toBe(false);
      expect(sendResponse).toHaveBeenCalledWith({ error: 'Unknown message type' });
    });
  });
});

describe('Background Service Worker with DEBUG_MODE', () => {
  let messageListener: Function;
  let consoleSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(async () => {
    vi.resetModules();
    vi.clearAllMocks();

    // Enable DEBUG_MODE
    vi.doMock('@/config/constants', () => ({
      DEBUG_MODE: true,
    }));

    vi.doMock('@/background/api-client', () => ({
      searchNutrition: vi.fn().mockResolvedValue({ data: null, cached: false }),
      clearCache: vi.fn().mockResolvedValue(undefined),
      getCacheStats: vi.fn().mockResolvedValue({ count: 0, keys: [] }),
    }));

    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

    (global.chrome.runtime.onMessage.addListener as ReturnType<typeof vi.fn>).mockImplementation(
      (listener: Function) => {
        messageListener = listener;
      }
    );

    await import('@/background/index');
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.resetModules();
  });

  it('should log messages when DEBUG_MODE is enabled', () => {
    expect(consoleSpy).toHaveBeenCalledWith('[MacroLens Background]', 'Background service worker initialized');
  });

  it('should log received messages when DEBUG_MODE is enabled', () => {
    const sendResponse = vi.fn();
    const message = { type: 'UNKNOWN_TYPE' };

    messageListener(message, {}, sendResponse);

    expect(consoleSpy).toHaveBeenCalledWith('[MacroLens Background]', 'Received message:', 'UNKNOWN_TYPE');
  });
});

