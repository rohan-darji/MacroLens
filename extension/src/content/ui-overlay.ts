// ui-overlay.ts - UI overlay for displaying nutrition information

import type { NutritionData } from '@/types/nutrition';
import { UI, DEBUG_MODE, CONFIDENCE } from '@/config/constants';

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
      <p>Analyzing nutrition data...</p>
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
      <div class="macrolens-error-icon">!</div>
      <h3>Unable to Load</h3>
      <p>${escapeHtml(message)}</p>
      <button class="macrolens-btn macrolens-btn-secondary macrolens-close-btn">Dismiss</button>
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
 * Calculates the stroke-dashoffset for a given calorie percentage
 * Based on a 2000 calorie daily value
 */
function getCalorieRingOffset(calories: number): number {
  const circumference = 314; // 2 * PI * 50 (radius)
  const percentage = Math.min(calories / 500, 1); // Cap at 500 cal (25% of 2000)
  return circumference * (1 - percentage);
}

/**
 * Shows nutrition data in the overlay
 */
export function showNutrition(data: NutritionData, cached: boolean = false, overlay?: HTMLElement): void {
  const element = overlay || document.getElementById(UI.OVERLAY_ID);
  if (!element) return;

  const confidenceClass = getConfidenceClass(data.confidence);
  const confidenceLabel = getConfidenceLabel(data.confidence);
  const ringOffset = getCalorieRingOffset(data.nutrients.calories);

  const warningHtml = data.confidence < CONFIDENCE.MEDIUM ? `
    <div class="macrolens-warning">
      Low confidence match — please verify manually
    </div>
  ` : '';

  element.innerHTML = `
    <div class="macrolens-content">
      <div class="macrolens-header">
        <div class="macrolens-brand">
          <div class="macrolens-logo">🔬</div>
          <h3 class="macrolens-title">MacroLens</h3>
        </div>
        <button class="macrolens-close-btn" title="Close">×</button>
      </div>

      <div class="macrolens-product">
        <div class="macrolens-product-name">${escapeHtml(data.productName)}</div>
        <div class="macrolens-serving">
          ${data.servingSize} ${escapeHtml(data.servingSizeUnit)}
        </div>
      </div>

      ${warningHtml}

      <div class="macrolens-nutrients">
        <div class="macrolens-calories-hero">
          <div class="macrolens-calories-ring">
            <svg viewBox="0 0 120 120">
              <defs>
                <linearGradient id="calorieGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" stop-color="#059669" />
                  <stop offset="100%" stop-color="#10B981" />
                </linearGradient>
              </defs>
              <circle class="ring-bg" cx="60" cy="60" r="50" />
              <circle class="ring-fill" cx="60" cy="60" r="50"
                style="stroke-dashoffset: ${ringOffset}" />
            </svg>
            <div class="macrolens-calories-value">
              <span class="macrolens-calories-number">${data.nutrients.calories.toFixed(0)}</span>
              <span class="macrolens-calories-unit">Calories</span>
            </div>
          </div>
        </div>

        <div class="macrolens-macros-grid">
          <div class="macrolens-macro-card protein macrolens-slide-up macrolens-stagger-1">
            <div class="macrolens-macro-icon">💪</div>
            <div class="macrolens-macro-value">${data.nutrients.protein.toFixed(1)}<span>g</span></div>
            <div class="macrolens-macro-label">Protein</div>
          </div>
          <div class="macrolens-macro-card carbs macrolens-slide-up macrolens-stagger-2">
            <div class="macrolens-macro-icon">⚡</div>
            <div class="macrolens-macro-value">${data.nutrients.carbohydrates.toFixed(1)}<span>g</span></div>
            <div class="macrolens-macro-label">Carbs</div>
          </div>
          <div class="macrolens-macro-card fat macrolens-slide-up macrolens-stagger-3">
            <div class="macrolens-macro-icon">🫒</div>
            <div class="macrolens-macro-value">${data.nutrients.totalFat.toFixed(1)}<span>g</span></div>
            <div class="macrolens-macro-label">Fat</div>
          </div>
        </div>
      </div>

      <div class="macrolens-footer">
        <div class="macrolens-confidence ${confidenceClass}">
          ${confidenceLabel}
        </div>
        <div class="macrolens-source">
          <span class="macrolens-source-badge">${cached ? 'Cached' : 'USDA'}</span>
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
  if (confidence >= CONFIDENCE.HIGH) return 'macrolens-confidence-high';
  if (confidence >= CONFIDENCE.MEDIUM) return 'macrolens-confidence-medium';
  return 'macrolens-confidence-low';
}

/**
 * Gets human-readable confidence label
 */
function getConfidenceLabel(confidence: number): string {
  if (confidence >= CONFIDENCE.HIGH) return 'High Match';
  if (confidence >= CONFIDENCE.MEDIUM) return 'Good Match';
  return 'Low Match';
}

/**
 * Escapes HTML to prevent XSS
 * Handles all characters significant in HTML and attribute contexts:
 * &, <, >, ", and '
 */
function escapeHtml(text: string): string {
  // Escape characters that are significant in HTML and attribute contexts
  const str = String(text);
  const map: { [key: string]: string } = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  };
  return str.replace(/[&<>"']/g, (ch) => map[ch]);
}
