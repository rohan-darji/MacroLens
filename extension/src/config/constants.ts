// constants.ts - Configuration constants

// API Configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
export const DEBUG_MODE = import.meta.env.VITE_DEBUG_MODE === 'true';

// Cache Configuration
export const CACHE_TTL_DAYS = parseInt(import.meta.env.VITE_CACHE_TTL_DAYS || '7', 10);
export const CACHE_TTL_MS = CACHE_TTL_DAYS * 24 * 60 * 60 * 1000;

// API Endpoints
export const ENDPOINTS = {
  HEALTH: '/health',
  NUTRITION_SEARCH: '/api/v1/nutrition/search',
} as const;

// Walmart Configuration
export const WALMART = {
  PRODUCT_URL_PATTERN: /^https:\/\/www\.walmart\.com\/ip\/[^\/?#]+\/\d+(?:[?#]|$)/,
  PRODUCT_PAGE_INDICATOR: '/ip/',
} as const;

// UI Configuration
export const UI = {
  OVERLAY_ID: 'macrolens-overlay',
  OVERLAY_CLASS: 'macrolens-overlay',
} as const;
