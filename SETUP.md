# MacroLens - Complete Setup Guide

This guide will walk you through setting up the MacroLens Chrome extension from scratch.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Backend Setup](#backend-setup)
4. [Extension Setup](#extension-setup)
5. [Loading Extension in Chrome](#loading-extension-in-chrome)
6. [Testing](#testing)
7. [Development Workflow](#development-workflow)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before starting, ensure you have the following installed:

### Required Software
- **Go**: Version 1.21 or higher ([Download](https://go.dev/dl/))
- **Node.js**: Version 18 or higher ([Download](https://nodejs.org/))
- **npm**: Comes with Node.js (verify with `npm --version`)
- **Chrome Browser**: Latest version
- **Git**: For cloning the repository

### Verify Installations

```bash
# Check Go version
go version
# Expected: go version go1.21.x or higher

# Check Node.js version
node --version
# Expected: v18.x.x or higher

# Check npm version
npm --version
# Expected: 9.x.x or higher
```

---

## Initial Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd MacroLens
```

### 2. Get USDA API Key (Required)

The backend requires a USDA FoodData Central API key:

1. Go to https://fdc.nal.usda.gov/api-key-signup/
2. Fill out the signup form
3. Check your email for the API key
4. Save this key - you'll need it in the next step

**Important**: Keep your API key private! Never commit it to Git.

---

## Backend Setup

### 1. Navigate to Backend Directory

```bash
cd backend
```

### 2. Create Environment File

Copy the example environment file:

```bash
# Windows
copy .env.example .env

# macOS/Linux
cp .env.example .env
```

### 3. Configure Environment Variables

Open `backend/.env` in your text editor and update:

```env
# Server Configuration
MACROLENS_SERVER_PORT=8080
MACROLENS_SERVER_ENVIRONMENT=development
MACROLENS_SERVER_ALLOWED_ORIGINS=chrome-extension://*

# USDA API Configuration
MACROLENS_USDA_API_KEY=YOUR_ACTUAL_API_KEY_HERE  # ← Replace this!
MACROLENS_USDA_BASE_URL=https://api.nal.usda.gov/fdc

# Cache Configuration
MACROLENS_CACHE_TYPE=memory
MACROLENS_CACHE_TTL=720h

# Rate Limiting
MACROLENS_RATELIMIT_PER_IP=100
MACROLENS_RATELIMIT_USDA=1000
```

**⚠️ IMPORTANT**: Replace `YOUR_ACTUAL_API_KEY_HERE` with your actual USDA API key from step 2!

### 4. Install Dependencies

Go will automatically download dependencies when you build/run:

```bash
go mod download
```

### 5. Build the Backend (Optional - verify it compiles)

```bash
go build -o bin/server ./cmd/server/main.go
```

If successful, you'll see no errors. You can delete the `bin/` folder after verifying.

### 6. Run the Backend Server

```bash
go run cmd/server/main.go
```

**Expected Output:**
```
Starting MacroLens Backend v1.0.0
Environment: development
Port: 8080
Cache Type: memory
Server listening on :8080
```

### 7. Test the Backend

Open a **new terminal** and test the health endpoint:

```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{"status":"healthy","service":"macrolens-backend","version":"1.0.0"}
```

**Backend is running!** Keep this terminal open.

---

## Extension Setup

Open a **new terminal** for the extension setup.

### 1. Navigate to Extension Directory

```bash
cd extension
```

(From project root: `cd extension`)

### 2. Install Dependencies

```bash
npm install
```

This will install all required packages (TypeScript, Vite, etc.)

**Expected Output:**
```
added 245 packages, and audited 246 packages in 15s
```

### 3. Build the Extension (Development Mode)

```bash
npm run dev
```

**Note**: This command will keep running and watch for file changes. The script automatically copies popup.html to the correct location.

**Expected Output:**
```
> macrolens-extension@1.0.0 dev
> node scripts/dev-watch.js

Starting Vite in watch mode...

vite v6.4.1 building for development...

watching for file changes...

build started...
transforming...
✓ 8 modules transformed.
rendering chunks...
computing gzip size...
dist/src/popup/index.html  0.86 kB │ gzip: 0.44 kB
dist/popup.css             0.66 kB │ gzip: 0.38 kB
dist/messages.js           0.17 kB │ gzip: 0.13 kB
dist/content.js            0.54 kB │ gzip: 0.35 kB
dist/background.js         0.55 kB │ gzip: 0.34 kB
dist/popup.js              1.12 kB │ gzip: 0.61 kB
built in 118ms.

Watching for popup.html changes...

✓ Copied popup.html to dist/popup.html
```

### 4. Verify Build Output

Check that the `dist/` folder contains:

```bash
# Windows
dir dist

# macOS/Linux
ls -la dist
```

**Expected files:**
```
dist/
├── background.js       ✓
├── content.js          ✓
├── content.css         ✓
├── manifest.json       ✓
├── popup.html          ✓ (copied in step 4)
├── popup.js            ✓
├── popup.css           ✓
├── messages.js         ✓
└── icons/
    ├── icon-16.png     ✓
    ├── icon-48.png     ✓
    └── icon-128.png    ✓
```

**Extension is built!**

---

## Loading Extension in Chrome

### 1. Open Chrome Extensions Page

In Chrome, navigate to:
```
chrome://extensions/
```

Or: **Menu (⋮)** → **Extensions** → **Manage Extensions**

### 2. Enable Developer Mode

Toggle **"Developer mode"** switch in the **top-right corner**.

![Developer Mode Toggle](https://developer.chrome.com/static/docs/extensions/get-started/tutorial/hello-world/image/extensions-page-e0d64d89a6acf_1920.png)

### 3. Load Unpacked Extension

1. Click **"Load unpacked"** button
2. Navigate to your project folder: `MacroLens/extension/dist`
3. Select the `dist` folder and click **"Select Folder"** (or "Open")

### 4. Verify Extension is Loaded

You should see:

**MacroLens Extension Card:**
```
┌─────────────────────────────────────┐
│ 🔵 MacroLens                        │
│ Version 1.0.0                       │
│ ID: <random-chrome-extension-id>    │
│ Instantly view nutritional...       │
│ ✓ Enabled                           │
│ 🔄 (Reload icon)  🗑️ (Remove)       │
└─────────────────────────────────────┘
```

**Extension Icon in Toolbar:**
- Look for the MacroLens icon (M logo) in your Chrome toolbar
- If not visible, click the puzzle piece icon 🧩 and pin MacroLens

**Extension is loaded!**

---

## Testing

### Test 1: Extension Activation

1. **Navigate to a Walmart Product Page**

   Example: https://www.walmart.com/ip/Great-Value-Whole-Milk-Gallon-128-Fl-Oz/10450114

2. **Open Chrome DevTools**

   Press **F12** or **Right-click** → **Inspect**

3. **Go to Console Tab**

4. **Look for MacroLens Logs**

   You should see:
   ```
   [MacroLens] Content script loaded
   [MacroLens] Walmart product page detected! https://www.walmart.com/ip/...
   [MacroLens] MacroLens extension is active
   ```

   **Content script is working!**

### Test 2: Background Service Worker

1. **Go to** `chrome://extensions/`

2. **Find MacroLens** and click **"service worker"** (blue link)

3. **Check Console Output**

   You should see:
   ```
   [MacroLens Background] Background service worker initialized
   ```

   **Background script is working!**

### Test 3: Extension Popup

1. **Click the MacroLens icon** in Chrome toolbar

2. **Verify Popup Displays:**
   ```
   MacroLens
   Nutrition information at your fingertips

   Status: Active
   Version: 1.0.0

   Visit a Walmart product page to see nutrition information.

   [Clear Cache]
   ```

   **Popup is working!**

### Test 4: Backend Health Check

In a terminal:

```bash
curl http://localhost:8080/health
```

**Expected:**
```json
{"status":"healthy","service":"macrolens-backend","version":"1.0.0"}
```

**Backend is healthy!**

---

## Development Workflow

### Starting Development

**Every time you want to work on MacroLens:**

1. **Terminal 1 - Backend:**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. **Terminal 2 - Extension (watch mode with auto-copy):**
   ```bash
   cd extension
   npm run dev
   ```

   The script automatically copies popup.html to the correct location on every build.

3. **Reload Extension in Chrome:**
   - After code changes, go to `chrome://extensions/`
   - Click the **reload icon (🔄)** on MacroLens card
   - Refresh the Walmart page you're testing on

### Making Changes

**Backend Changes:**
1. Edit Go files in `backend/`
2. Save the file
3. **Stop the server** (Ctrl+C in Terminal 1)
4. **Restart:** `go run cmd/server/main.go`

**Extension Changes:**
1. Edit TypeScript files in `extension/src/`
2. Save the file (Vite auto-rebuilds in watch mode)
3. **Reload extension** in Chrome (`chrome://extensions/` → 🔄)
4. **Refresh the test page**

### Building for Production

**Extension:**
```bash
cd extension
npm run build
```

Then copy popup.html as before.

**Backend:**
```bash
cd backend
go build -o bin/server ./cmd/server/main.go
./bin/server
```

---

## Troubleshooting

### Issue: Backend won't start - "USDA API key is required"

**Solution:**
- Check `backend/.env` file exists
- Verify `MACROLENS_USDA_API_KEY` is set to your actual API key (not `DEMO_KEY`)
- Make sure there are no spaces around the `=` sign

### Issue: Extension shows "ERR_FILE_NOT_FOUND"

**Solution:**
- Make sure you copied `popup.html` to `dist/` folder:
  ```bash
  cp dist/src/popup/index.html dist/popup.html
  ```
- Reload the extension in Chrome

### Issue: No console logs from extension

**Solution:**
- Make sure you built in **development mode**: `npm run dev`
- Production builds (`npm run build`) disable debug logs
- Check the correct page: `chrome://extensions/` for background logs, Walmart page for content logs

### Issue: Extension doesn't detect Walmart page

**Solution:**
- URL must match: `https://www.walmart.com/ip/*`
- Try this test URL: https://www.walmart.com/ip/Great-Value-Whole-Milk-Gallon-128-Fl-Oz/10450114
- Check Console for any errors

### Issue: npm install fails

**Solution:**
- Check Node.js version: `node --version` (must be 18+)
- Delete `node_modules` and try again:
  ```bash
  rm -rf node_modules package-lock.json
  npm install
  ```

### Issue: Go dependencies fail

**Solution:**
- Check Go version: `go version` (must be 1.21+)
- Clear Go cache:
  ```bash
  go clean -modcache
  go mod download
  ```

### Issue: Extension changes not reflecting

**Solution:**
1. Make sure `npm run dev` is running (watch mode)
2. Reload extension: `chrome://extensions/` → 🔄
3. **Hard refresh** the test page: Ctrl+Shift+R (Windows) or Cmd+Shift+R (Mac)
4. Clear browser cache if needed

### Issue: Port 8080 already in use

**Solution:**
- Find and kill the process using port 8080:
  ```bash
  # Windows
  netstat -ano | findstr :8080
  taskkill /PID <PID> /F

  # macOS/Linux
  lsof -ti:8080 | xargs kill -9
  ```
- Or change the port in `backend/.env`:
  ```env
  MACROLENS_SERVER_PORT=3000
  ```
  And update `extension/.env.development`:
  ```env
  VITE_API_BASE_URL=http://localhost:3000
  ```

---

## Next Steps

**Phase 1 Complete!** Your setup is working.

**Phase 2** will implement:
- Walmart product name extraction
- USDA API nutrition lookup
- UI overlay with nutrition display
- Complete end-to-end data flow

**Resources:**
- [Chrome Extension Documentation](https://developer.chrome.com/docs/extensions/)
- [USDA FoodData Central API](https://fdc.nal.usda.gov/api-guide.html)
- [Project README](README.md)

---

## Running Tests

MacroLens has comprehensive unit tests for Phase 1 functionality.

### Backend Tests (Go)

```bash
cd backend

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/infrastructure/usda
go test ./internal/infrastructure/cache
go test ./internal/delivery/http
```

**Test Coverage:**
- USDA data mapper (6 tests)
- In-memory cache (7 test suites)
- HTTP middleware (3 test suites)
- **Total: 26 test cases**

### Extension Tests (TypeScript)

```bash
cd extension

# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage report
npm run test:coverage

# Run tests with interactive UI
npm run test:ui
```

**Test Coverage:**
- Walmart page detection (14 tests)
- Message types (10 tests)
- Configuration constants (15 tests)
- **Total: 39 test cases**

---

## Quick Reference Commands

```bash
# Backend
cd backend
go run cmd/server/main.go          # Run server
curl http://localhost:8080/health  # Test health endpoint
go test ./...                      # Run tests

# Extension
cd extension
npm install                        # Install dependencies (once)
npm run dev                        # Development build (watch mode, auto-copies popup.html)
npm run build                      # Production build (auto-copies popup.html)
npm run type-check                 # Check TypeScript types
npm test                           # Run tests
```