import { build } from 'vite';
import { resolve } from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = resolve(__dirname, '..');

// Environment configuration for build
const isDev = process.env.NODE_ENV === 'development';
const envDefine = {
  'import.meta.env.VITE_API_BASE_URL': JSON.stringify(
    process.env.VITE_API_BASE_URL || 'http://localhost:8080'
  ),
  'import.meta.env.VITE_DEBUG_MODE': JSON.stringify(
    process.env.VITE_DEBUG_MODE || (isDev ? 'true' : 'false')
  ),
  'import.meta.env.VITE_CACHE_TTL_DAYS': JSON.stringify(
    process.env.VITE_CACHE_TTL_DAYS || '7'
  ),
};

// Build content script as IIFE (no imports)
async function buildContent() {
  await build({
    configFile: false, // Don't load vite.config.ts
    root,
    define: envDefine,
    build: {
      lib: {
        entry: resolve(root, 'src/content/index.ts'),
        name: 'MacroLensContent',
        fileName: () => 'content.js',
        formats: ['iife'],
      },
      rollupOptions: {
        output: {
          // Force all code into single file - no code splitting
          inlineDynamicImports: true,
        },
      },
      outDir: resolve(root, 'dist'),
      emptyOutDir: false,
      sourcemap: process.env.NODE_ENV === 'development',
    },
    resolve: {
      alias: {
        '@': resolve(root, 'src'),
      },
    },
  });
}

// Build background script as ES module
async function buildBackground() {
  await build({
    configFile: false, // Don't load vite.config.ts
    root,
    define: envDefine,
    build: {
      lib: {
        entry: resolve(root, 'src/background/index.ts'),
        name: 'MacroLensBackground',
        fileName: () => 'background.js',
        formats: ['es'],
      },
      rollupOptions: {
        output: {
          inlineDynamicImports: true,
        },
      },
      outDir: resolve(root, 'dist'),
      emptyOutDir: false,
      sourcemap: process.env.NODE_ENV === 'development',
    },
    resolve: {
      alias: {
        '@': resolve(root, 'src'),
      },
    },
  });
}

// Build popup script as ES module + copy public assets
async function buildPopup() {
  await build({
    configFile: false, // Don't load vite.config.ts
    root,
    publicDir: 'public', // Relative to root - copy public dir (manifest, icons, CSS)
    define: envDefine,
    build: {
      lib: {
        entry: resolve(root, 'src/popup/index.ts'),
        name: 'MacroLensPopup',
        fileName: () => 'popup.js',
        formats: ['es'],
      },
      rollupOptions: {
        output: {
          inlineDynamicImports: true,
        },
      },
      outDir: resolve(root, 'dist'),
      emptyOutDir: true, // Only the first build empties
      sourcemap: process.env.NODE_ENV === 'development',
      copyPublicDir: true, // Ensure public dir is copied
    },
    resolve: {
      alias: {
        '@': resolve(root, 'src'),
      },
    },
  });
}

// Run builds in sequence
console.log('Building MacroLens extension...');
await buildPopup();
console.log('✓ Popup built');
await buildContent();
console.log('✓ Content script built');
await buildBackground();
console.log('✓ Background script built');
console.log('Build complete!');
