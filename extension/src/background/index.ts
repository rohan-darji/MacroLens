// background/index.ts - Background service worker

import { MessageType } from '@/types/messages';
import { DEBUG_MODE } from '@/config/constants';

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
      // TODO: Implement in Phase 2
      log('GET_NUTRITION request received (not yet implemented)');
      sendResponse({ error: 'Not yet implemented - coming in Phase 2' });
      break;

    case MessageType.CLEAR_CACHE:
      // TODO: Implement in Phase 2
      log('CLEAR_CACHE request received (not yet implemented)');
      sendResponse({ success: true });
      break;

    default:
      log('Unknown message type:', message.type);
      sendResponse({ error: 'Unknown message type' });
  }

  return true; // Keep message channel open for async response
});

log('Background service worker initialized');
