import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { NutritionData } from '@/types/nutrition';

// Mock the constants
vi.mock('@/config/constants', () => ({
  UI: {
    OVERLAY_ID: 'macrolens-overlay',
    OVERLAY_CLASS: 'macrolens-overlay',
  },
  DEBUG_MODE: false,
  CONFIDENCE: {
    HIGH: 90,
    MEDIUM: 70,
  },
}));

describe('UI Overlay', () => {
  beforeEach(() => {
    // Clean up any existing overlays
    const existing = document.getElementById('macrolens-overlay');
    if (existing) {
      existing.remove();
    }
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
  });

  describe('createOverlay', () => {
    it('should create an overlay element with correct id and class', async () => {
      const { createOverlay } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      expect(overlay.id).toBe('macrolens-overlay');
      expect(overlay.className).toBe('macrolens-overlay');
    });

    it('should append overlay to document body', async () => {
      const { createOverlay } = await import('@/content/ui-overlay');
      createOverlay();

      const element = document.getElementById('macrolens-overlay');
      expect(element).not.toBeNull();
    });

    it('should remove existing overlay before creating new one', async () => {
      const { createOverlay } = await import('@/content/ui-overlay');

      // Create first overlay
      createOverlay();

      // Create second overlay
      const secondOverlay = createOverlay();

      // Should only have one overlay
      const overlays = document.querySelectorAll('#macrolens-overlay');
      expect(overlays.length).toBe(1);
      expect(overlays[0]).toBe(secondOverlay);
    });

    it('should show loading state initially', async () => {
      const { createOverlay } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      expect(overlay.innerHTML).toContain('macrolens-loading');
      expect(overlay.innerHTML).toContain('macrolens-spinner');
    });
  });

  describe('removeOverlay', () => {
    it('should remove existing overlay', async () => {
      const { createOverlay, removeOverlay } = await import('@/content/ui-overlay');

      createOverlay();
      expect(document.getElementById('macrolens-overlay')).not.toBeNull();

      removeOverlay();
      expect(document.getElementById('macrolens-overlay')).toBeNull();
    });

    it('should not throw if overlay does not exist', async () => {
      const { removeOverlay } = await import('@/content/ui-overlay');
      expect(() => removeOverlay()).not.toThrow();
    });
  });

  describe('showLoading', () => {
    it('should display loading spinner in overlay', async () => {
      const { createOverlay, showLoading } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      showLoading(overlay);

      expect(overlay.innerHTML).toContain('macrolens-loading');
      expect(overlay.innerHTML).toContain('macrolens-spinner');
      expect(overlay.innerHTML).toContain('Analyzing nutrition data...');
    });

    it('should find overlay by ID if not provided', async () => {
      const { createOverlay, showLoading } = await import('@/content/ui-overlay');
      createOverlay();

      showLoading();

      const overlay = document.getElementById('macrolens-overlay');
      expect(overlay?.innerHTML).toContain('macrolens-loading');
    });
  });

  describe('showError', () => {
    it('should display error message in overlay', async () => {
      const { createOverlay, showError } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      showError('Test error message', overlay);

      expect(overlay.innerHTML).toContain('macrolens-error');
      expect(overlay.innerHTML).toContain('Unable to Load');
      expect(overlay.innerHTML).toContain('Test error message');
    });

    it('should include close button', async () => {
      const { createOverlay, showError } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      showError('Error', overlay);

      const closeBtn = overlay.querySelector('.macrolens-close-btn');
      expect(closeBtn).not.toBeNull();
    });
  });

  describe('showNutrition', () => {
    const mockNutritionData: NutritionData = {
      fdcId: 123456,
      productName: 'Test Product',
      brand: 'Test Brand',
      servingSize: 240,
      servingSizeUnit: 'ml',
      confidence: 85,
      source: 'USDA',
      nutrients: {
        calories: 150,
        protein: 8,
        carbohydrates: 12,
        totalFat: 8,
        saturatedFat: 5,
        fiber: 0,
        sugar: 12,
        sodium: 130,
      },
    };

    it('should display nutrition data in overlay', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      showNutrition(mockNutritionData, false, overlay);

      expect(overlay.innerHTML).toContain('MacroLens');
      expect(overlay.innerHTML).toContain('Test Product');
      expect(overlay.innerHTML).toContain('150'); // calories
      expect(overlay.innerHTML).toContain('8.0'); // protein value
    });

    it('should show cached indicator when data is from cache', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      showNutrition(mockNutritionData, true, overlay);

      expect(overlay.innerHTML).toContain('Cached');
    });

    it('should show USDA indicator when data is not cached', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      showNutrition(mockNutritionData, false, overlay);

      expect(overlay.innerHTML).toContain('USDA');
    });

    it('should display confidence label', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();

      showNutrition(mockNutritionData, false, overlay);

      // 85% confidence shows "Good Match" label
      expect(overlay.innerHTML).toContain('Good Match');
    });

    it('should show warning for low confidence matches', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();
      const lowConfidenceData = { ...mockNutritionData, confidence: 50 };

      showNutrition(lowConfidenceData, false, overlay);

      expect(overlay.innerHTML).toContain('macrolens-warning');
      expect(overlay.innerHTML).toContain('Low confidence match');
    });

    it('should not show warning for high confidence matches', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();
      const highConfidenceData = { ...mockNutritionData, confidence: 95 };

      showNutrition(highConfidenceData, false, overlay);

      expect(overlay.innerHTML).not.toContain('macrolens-warning');
    });

    it('should apply high confidence class for 90%+', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();
      const highConfidenceData = { ...mockNutritionData, confidence: 95 };

      showNutrition(highConfidenceData, false, overlay);

      expect(overlay.innerHTML).toContain('macrolens-confidence-high');
    });

    it('should apply medium confidence class for 70-89%', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();
      const mediumConfidenceData = { ...mockNutritionData, confidence: 75 };

      showNutrition(mediumConfidenceData, false, overlay);

      expect(overlay.innerHTML).toContain('macrolens-confidence-medium');
    });

    it('should apply low confidence class for below 70%', async () => {
      const { createOverlay, showNutrition } = await import('@/content/ui-overlay');
      const overlay = createOverlay();
      const lowConfidenceData = { ...mockNutritionData, confidence: 50 };

      showNutrition(lowConfidenceData, false, overlay);

      expect(overlay.innerHTML).toContain('macrolens-confidence-low');
    });
  });
});
