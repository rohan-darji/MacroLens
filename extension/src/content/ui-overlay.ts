// ui-overlay.ts - UI overlay for displaying nutrition information

import type { NutritionData } from '@/types/nutrition';
import { UI, DEBUG_MODE } from '@/config/constants';

/**
 * Logs debug messages if DEBUG_MODE is enabled
 */
function log(...args: any[]): void {
  if (DEBUG_MODE) {
    console.log('[MacroLens UI]', ...args);
  }
}

/**
 * Creates and injects the nutrition overlay into the page
 */
export function createOverlay(): HTMLElement {
  // Remove existing overlay if present
  removeOverlay();

  const overlay = document.createElement('div');
  overlay.id = UI.OVERLAY_ID;
  overlay.className = UI.OVERLAY_CLASS;

  // Show loading state initially
  showLoading(overlay);

  document.body.appendChild(overlay);
  log('Overlay created');

  return overlay;
}

/**
 * Removes the overlay from the page
 */
export function removeOverlay(): void {
  const existing = document.getElementById(UI.OVERLAY_ID);
  if (existing) {
    existing.remove();
    log('Overlay removed');
  }
}

/**
 * Shows loading state in the overlay
 */
export function showLoading(overlay?: HTMLElement): void {
  const element = overlay || document.getElementById(UI.OVERLAY_ID);
  if (!element) return;

  element.innerHTML = `
    <div class="macrolens-loading">
      <div class="macrolens-spinner"></div>
      <p>Loading nutrition info...</p>
    </div>
  `;
}

/**
 * Shows error message in the overlay
 */
export function showError(message: string, overlay?: HTMLElement): void {
  const element = overlay || document.getElementById(UI.OVERLAY_ID);
  if (!element) return;

  element.innerHTML = `
    <div class="macrolens-error">
      <h3>Error</h3>
      <p>${escapeHtml(message)}</p>
      <button class="macrolens-close-btn">Close</button>
    </div>
  `;

  // Add close button handler
  const closeBtn = element.querySelector('.macrolens-close-btn');
  if (closeBtn) {
    closeBtn.addEventListener('click', removeOverlay);
  }

  log('Error shown:', message);
}

/**
 * Shows nutrition data in the overlay
 */
export function showNutrition(data: NutritionData, cached: boolean = false, overlay?: HTMLElement): void {
  const element = overlay || document.getElementById(UI.OVERLAY_ID);
  if (!element) return;

  const confidenceClass = getConfidenceClass(data.confidence);
  const warningHtml = data.confidence < 70 ? `
    <div class="macrolens-warning">
      Low confidence match - verify manually
    </div>
  ` : '';

  element.innerHTML = `
    <div class="macrolens-content">
      <div class="macrolens-header">
        <h3 class="macrolens-title">Nutrition Info</h3>
        <button class="macrolens-close-btn" title="Close">×</button>
      </div>

      <div class="macrolens-product">
        <div class="macrolens-product-name">${escapeHtml(data.productName)}</div>
        <div class="macrolens-serving">
          Serving: ${escapeHtml(data.servingSize)} ${escapeHtml(data.servingSizeUnit)}
        </div>
      </div>

      ${warningHtml}

      <div class="macrolens-nutrients">
        <div class="macrolens-nutrient-row macrolens-calories">
          <span class="macrolens-nutrient-label">Calories</span>
          <span class="macrolens-nutrient-value">${data.nutrients.calories.toFixed(0)}</span>
        </div>
        <div class="macrolens-nutrient-row">
          <span class="macrolens-nutrient-label">Protein</span>
          <span class="macrolens-nutrient-value">${data.nutrients.protein.toFixed(1)}g</span>
        </div>
        <div class="macrolens-nutrient-row">
          <span class="macrolens-nutrient-label">Carbohydrates</span>
          <span class="macrolens-nutrient-value">${data.nutrients.carbohydrates.toFixed(1)}g</span>
        </div>
        <div class="macrolens-nutrient-row">
          <span class="macrolens-nutrient-label">Total Fat</span>
          <span class="macrolens-nutrient-value">${data.nutrients.totalFat.toFixed(1)}g</span>
        </div>
      </div>

      <div class="macrolens-footer">
        <div class="macrolens-confidence ${confidenceClass}">
          Match: ${data.confidence.toFixed(0)}%
        </div>
        <div class="macrolens-source">
          ${cached ? 'Cached' : 'USDA'}
        </div>
      </div>
    </div>
  `;

  // Add close button handler
  const closeBtn = element.querySelector('.macrolens-close-btn');
  if (closeBtn) {
    closeBtn.addEventListener('click', removeOverlay);
  }

  log('Nutrition data displayed', { cached, confidence: data.confidence });
}

/**
 * Gets CSS class based on confidence score
 */
function getConfidenceClass(confidence: number): string {
  if (confidence >= 90) return 'macrolens-confidence-high';
  if (confidence >= 70) return 'macrolens-confidence-medium';
  return 'macrolens-confidence-low';
}

/**
 * Escapes HTML to prevent XSS
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
