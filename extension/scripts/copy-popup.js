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

const htmlSource = join(__dirname, '../src/popup/index.html');
const htmlDest = join(__dirname, '../dist/popup.html');
const cssSource = join(__dirname, '../src/popup/popup.css');
const cssDest = join(__dirname, '../dist/popup.css');

try {
  // Copy popup.html
  if (existsSync(htmlSource)) {
    copyFileSync(htmlSource, htmlDest);
    console.log('✓ Copied popup.html to dist/popup.html');
  } else {
    console.warn('⚠ Warning: src/popup/index.html not found');
  }

  // Copy popup.css
  if (existsSync(cssSource)) {
    copyFileSync(cssSource, cssDest);
    console.log('✓ Copied popup.css to dist/popup.css');
  } else {
    console.warn('⚠ Warning: src/popup/popup.css not found');
  }
} catch (error) {
  console.error('✗ Failed to copy popup files:', error.message);
  process.exit(1);
}
