#!/usr/bin/env node
/**
 * Development script that watches for changes and rebuilds
 */

import { watch as fsWatch } from 'chokidar';
import { execSync } from 'child_process';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const srcDir = join(__dirname, '../src');
const publicDir = join(__dirname, '../public');

console.log('🔨 Building extension...\n');

// Initial build
try {
  execSync('npm run build', { stdio: 'inherit', cwd: join(__dirname, '..') });
  console.log('\n✓ Initial build complete');
} catch (error) {
  console.error('✗ Build failed:', error.message);
}

console.log('\n👀 Watching for file changes...\n');

// Watch for changes
const watcher = fsWatch([srcDir, publicDir], {
  ignored: /(^|[\/\\])\../, // ignore dotfiles
  persistent: true,
  ignoreInitial: true,
});

let building = false;
let pendingRebuild = false;

watcher.on('change', (path) => {
  console.log(`📝 File changed: ${path}`);

  if (building) {
    pendingRebuild = true;
    return;
  }

  building = true;
  console.log('🔨 Rebuilding...');

  try {
    execSync('npm run build', { stdio: 'inherit', cwd: join(__dirname, '..') });
    console.log('✓ Build complete\n');
  } catch (error) {
    console.error('✗ Build failed:', error.message);
  } finally {
    building = false;

    if (pendingRebuild) {
      pendingRebuild = false;
      setTimeout(() => watcher.emit('change', 'pending'), 100);
    }
  }
});

// Handle cleanup on exit
process.on('SIGINT', () => {
  console.log('\n\nStopping development watcher...');
  watcher.close();
  process.exit(0);
});

process.on('SIGTERM', () => {
  watcher.close();
  process.exit(0);
});
