// popup/index.ts - Popup script with enhanced UI interactions

import { MessageType } from '@/types/messages';

/**
 * Shows a toast notification
 */
function showToast(message: string, type: 'success' | 'error' = 'success'): void {
  // Remove existing toast if any
  const existingToast = document.querySelector('.toast');
  if (existingToast) {
    existingToast.remove();
  }

  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.textContent = message;
  document.body.appendChild(toast);

  // Trigger animation
  requestAnimationFrame(() => {
    toast.classList.add('show');
  });

  // Remove after delay
  setTimeout(() => {
    toast.classList.remove('show');
    setTimeout(() => toast.remove(), 300);
  }, 2500);
}

/**
 * Updates the cache count display
 */
async function updateCacheCount(): Promise<void> {
  const cacheCountEl = document.getElementById('cache-count');
  if (!cacheCountEl) return;

  try {
    const storage = await chrome.storage.local.get(null);
    const cacheKeys = Object.keys(storage).filter(key => key.startsWith('nutrition:'));
    cacheCountEl.textContent = cacheKeys.length.toString();
  } catch (error) {
    cacheCountEl.textContent = '-';
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const clearCacheBtn = document.getElementById('clear-cache');

  // Update cache count on load
  updateCacheCount();

  clearCacheBtn?.addEventListener('click', async () => {
    try {
      // Add loading state
      clearCacheBtn.classList.add('loading');
      const originalHtml = clearCacheBtn.innerHTML;
      clearCacheBtn.innerHTML = '<span class="btn-icon">⏳</span> Clearing...';

      // Send message to background script to clear cache
      chrome.runtime.sendMessage(
        { type: MessageType.CLEAR_CACHE },
        (response) => {
          // Restore button state
          clearCacheBtn.classList.remove('loading');
          clearCacheBtn.innerHTML = originalHtml;

          if (response?.success) {
            // Show success feedback
            clearCacheBtn.classList.add('success');
            clearCacheBtn.innerHTML = '<span class="btn-icon">✓</span> Cleared!';
            showToast('Cache cleared successfully', 'success');

            // Update cache count
            updateCacheCount();

            // Reset button after delay
            setTimeout(() => {
              clearCacheBtn.classList.remove('success');
              clearCacheBtn.innerHTML = originalHtml;
            }, 2000);
          } else {
            showToast('Failed to clear cache', 'error');
          }
        }
      );
    } catch (error) {
      console.error('Error clearing cache:', error);
      showToast('Error clearing cache', 'error');
    }
  });
});
