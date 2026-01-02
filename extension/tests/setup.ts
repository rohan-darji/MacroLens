// Test setup file for Vitest

import { vi } from 'vitest';

// Mock Chrome APIs (not available in jsdom)
global.chrome = {
  runtime: {
    onMessage: {
      addListener: vi.fn(),
    },
    sendMessage: vi.fn(),
    lastError: null,
  },
  storage: {
    local: {
      get: vi.fn().mockResolvedValue({}),
      set: vi.fn().mockResolvedValue(undefined),
      remove: vi.fn().mockResolvedValue(undefined),
      clear: vi.fn().mockResolvedValue(undefined),
    },
  },
} as any;

// Note: Do NOT override document or window here - jsdom provides these
// Only mock Chrome-specific APIs that jsdom doesn't provide
