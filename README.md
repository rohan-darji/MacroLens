# MacroLens

MacroLens is a Chrome extension that instantly displays nutritional information for food products while shopping online. Browse Walmart product pages and see key macros (calories, protein, carbs, fat) without leaving the page.

## Features

- Automatic detection of Walmart product pages
- Display nutrition data: Calories, Protein, Carbs, Fat, Serving Size
- Two-tier caching (extension + backend) for fast performance
- Intelligent product matching with USDA FoodData Central
- Non-intrusive UI overlay
- <500ms response time (cached), <3s (API call)

## Tech Stack

- **Frontend**: TypeScript, Chrome Extension Manifest V3, Vite
- **Backend**: Go, Gin framework, Clean Architecture
- **Data Source**: USDA FoodData Central API
- **Cache**: In-memory (dev), Redis (production)

## Prerequisites

- Go 1.21+
- Node.js 18+
- Chrome browser
- USDA API Key ([Get one here](https://fdc.nal.usda.gov/api-key-signup/))

## Quick Start

**New to the project?** Follow the [Complete Setup Guide](SETUP.md)

**Already set up?** Quick commands:
```bash
# Terminal 1: Start backend
cd backend && go run cmd/server/main.go

# Terminal 2: Start extension watcher
cd extension && npm run dev

# Then reload extension in Chrome (chrome://extensions/)
```

## Project Structure

```
MacroLens/
├── backend/          # Go backend API
│   ├── cmd/         # Application entry point
│   ├── internal/    # Clean architecture layers
│   └── tests/       # Unit tests
├── extension/        # Chrome extension (TypeScript)
│   ├── src/         # Source code
│   ├── tests/       # Unit tests
│   └── dist/        # Build output (load in Chrome)
├── SETUP.md         # Complete setup guide
└── README.md        # This file
```

## Architecture

```
User → Walmart Product Page
  → Content Script (detect & extract)
  → Background Script (message handler)
  → Backend API (proxy & cache)
  → USDA FoodData Central API
  → Return nutrition data
  → Display in UI overlay
```

**Design Pattern**: Clean Architecture (Backend)
- Domain Layer: Core business entities
- Usecase Layer: Business logic
- Delivery Layer: HTTP handlers
- Infrastructure Layer: External services (USDA, cache)

## Testing

```bash
# Backend tests
cd backend && go test ./...

# Extension tests
cd extension && npm test
```

See [SETUP.md - Running Tests](SETUP.md#running-tests) for details.

## Documentation

- **[SETUP.md](SETUP.md)** - Complete setup guide

## License

Business Source License 1.1 (BSL 1.1)

**Summary**:
- **Allowed**: Personal use, educational use, internal testing, modifications, redistribution
- **Not Allowed**: Commercial use (selling, hosting as paid service, embedding in paid products)
- **Change Date**: 2030-12-28 (converts to Apache License 2.0)

See [LICENSE](LICENSE) for full terms.

## Acknowledgments

- Nutrition data provided by [USDA FoodData Central](https://fdc.nal.usda.gov/)
