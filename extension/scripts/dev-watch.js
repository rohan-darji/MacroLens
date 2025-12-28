#!/usr/bin/env node
/**
 * Development script that runs Vite in watch mode and copies popup.html on changes
 */

import { spawn } from 'child_process';
import { watch, copyFileSync, existsSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const source = join(__dirname, '../dist/src/popup/index.html');
const dest = join(__dirname, '../dist/popup.html');

// Function to copy popup.html
function copyPopup() {
  try {
    if (existsSync(source)) {
      copyFileSync(source, dest);
      console.log('✓ Copied popup.html to dist/popup.html');
    }
  } catch (error) {
    console.error('✗ Failed to copy popup.html:', error.message);
  }
}

// Start Vite in watch mode
console.log('Starting Vite in watch mode...\n');
const vite = spawn('npx', ['vite', 'build', '--watch', '--mode', 'development'], {
  stdio: 'inherit',
  shell: true
});

// Watch for changes to popup HTML and copy it
const watchDir = join(__dirname, '../dist/src/popup');
let watcher;

// Wait a bit for Vite to create the dist directory
setTimeout(() => {
  if (existsSync(watchDir)) {
    console.log('\n👀 Watching for popup.html changes...\n');
    watcher = watch(watchDir, (eventType, filename) => {
      if (filename === 'index.html') {
        copyPopup();
      }
    });

    // Copy initially if file exists
    copyPopup();
  }
}, 2000);

// Handle cleanup on exit
process.on('SIGINT', () => {
  console.log('\n\nStopping development server...');
  if (watcher) watcher.close();
  vite.kill();
  process.exit(0);
});

process.on('SIGTERM', () => {
  if (watcher) watcher.close();
  vite.kill();
  process.exit(0);
});
