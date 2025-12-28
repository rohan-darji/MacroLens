#!/usr/bin/env node
/**
 * Post-build script to copy popup.html to the correct location
 * Vite outputs popup HTML to dist/src/popup/index.html but manifest expects dist/popup.html
 */

import { copyFileSync, existsSync, mkdirSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const source = join(__dirname, '../dist/src/popup/index.html');
const dest = join(__dirname, '../dist/popup.html');

try {
  if (existsSync(source)) {
    copyFileSync(source, dest);
    console.log('✓ Copied popup.html to dist/popup.html');
  } else {
    console.warn('⚠ Warning: dist/src/popup/index.html not found');
  }
} catch (error) {
  console.error('✗ Failed to copy popup.html:', error.message);
  process.exit(1);
}
