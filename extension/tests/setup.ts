// Test setup file for Vitest

import { vi } from 'vitest';

// Mock Chrome APIs
global.chrome = {
  runtime: {
    onMessage: {
      addListener: vi.fn(),
    },
    sendMessage: vi.fn(),
  },
  storage: {
    local: {
      get: vi.fn(),
      set: vi.fn(),
      remove: vi.fn(),
      clear: vi.fn(),
    },
  },
} as any;

// Mock window.location for tests
delete (global as any).window;
(global as any).window = {
  location: {
    href: '',
  },
};

// Mock document for DOM tests
(global as any).document = {
  readyState: 'complete',
  addEventListener: vi.fn(),
  querySelector: vi.fn(),
  querySelectorAll: vi.fn(),
};

// Mock console methods if needed
global.console = {
  ...console,
  log: vi.fn(),
  error: vi.fn(),
  warn: vi.fn(),
};
