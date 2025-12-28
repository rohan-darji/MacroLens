// popup/index.ts - Popup script

import { MessageType } from '@/types/messages';

document.addEventListener('DOMContentLoaded', () => {
  const clearCacheBtn = document.getElementById('clear-cache');

  clearCacheBtn?.addEventListener('click', async () => {
    try {
      // Send message to background script to clear cache
      chrome.runtime.sendMessage(
        { type: MessageType.CLEAR_CACHE },
        (response) => {
          if (response?.success) {
            alert('Cache cleared successfully!');
          } else {
            alert('Failed to clear cache');
          }
        }
      );
    } catch (error) {
      console.error('Error clearing cache:', error);
      alert('Error clearing cache');
    }
  });
});
