// background/index.ts - Background service worker

import { MessageType, type GetNutritionPayload, type NutritionResponsePayload } from '@/types/messages';
import { DEBUG_MODE } from '@/config/constants';
import { searchNutrition, clearCache, getCacheStats } from './api-client';

/**
 * Logs debug messages if DEBUG_MODE is enabled
 */
function log(...args: any[]): void {
  if (DEBUG_MODE) {
    console.log('[MacroLens Background]', ...args);
  }
}

// Listen for messages from content scripts
chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  log('Received message:', message.type);

  switch (message.type) {
    case MessageType.GET_NUTRITION:
      handleGetNutrition(message.payload as GetNutritionPayload, sendResponse);
      return true; // Keep message channel open for async response

    case MessageType.CLEAR_CACHE:
      handleClearCache(sendResponse);
      return true; // Keep message channel open for async response

    default:
      log('Unknown message type:', message.type);
      sendResponse({ error: 'Unknown message type' });
      return false;
  }
});

/**
 * Handles GET_NUTRITION messages from content scripts
 */
async function handleGetNutrition(
  payload: GetNutritionPayload,
  sendResponse: (response: NutritionResponsePayload) => void
): Promise<void> {
  try {
    log('GET_NUTRITION request:', payload.productInfo);

    const result = await searchNutrition(payload.productInfo);

    const response: NutritionResponsePayload = {
      data: result.data,
      error: result.error,
      cached: result.cached,
    };

    sendResponse(response);
  } catch (error) {
    log('Error handling GET_NUTRITION:', error);
    sendResponse({
      error: error instanceof Error ? error.message : 'Unknown error occurred',
      cached: false,
    });
  }
}

/**
 * Handles CLEAR_CACHE messages
 */
async function handleClearCache(sendResponse: (response: any) => void): Promise<void> {
  try {
    log('CLEAR_CACHE request');

    await clearCache();
    const stats = await getCacheStats();

    sendResponse({
      success: true,
      message: 'Cache cleared successfully',
      stats,
    });
  } catch (error) {
    log('Error clearing cache:', error);
    sendResponse({
      success: false,
      error: error instanceof Error ? error.message : 'Failed to clear cache',
    });
  }
}

log('Background service worker initialized');
