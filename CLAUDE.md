# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

智账系统 (zhizhang-server) — A full-stack ERP-lite system with Go/Gin backend and Vue 3 frontend.

- **Backend**: Go 1.26 + Gin + GORM (PostgreSQL) + Redis + zerolog
- **Frontend**: Vue 3 + TypeScript + Vite + Element Plus + Pinia + Axios
- **Business domains**: Product, Purchase, Sale, Inventory, CRM, Mall, Approval, Price, Warehouse, Customer

## Common Commands

### Backend

```bash
# Build
make build                    # Compile binary to build/zhizhang-server
make run                      # Run without hot-reload
go run cmd/server/main.go     # Direct run

# Dev (requires `air` for hot-reload)
make dev

# Tests
go test -v -race -cover ./...          # All tests with race detection
go test -v -race ./...                 # All tests without cover
go test -v -race -run TestXxx ./...    # Single test (replace TestXxx)

# Lint
make lint                     # golangci-lint (auto-installs if missing)

# Database
make migrate                  # Run auto-migration via main.go --migrate
make pg-up                    # Start PostgreSQL container (port 5432)
make redis-up                 # Start Redis container (port 6379)
make infra-up                 # Start both PostgreSQL + Redis

# Swagger docs
make swagger                  # Generates docs/ from handler annotations
```

### Frontend

```bash
cd frontend
npm run dev                   # Vite dev server (port 5173)
npm run build                 # Production build
npm run preview               # Preview production build
```

## Architecture

### Backend (`internal/`)

The backend follows a **layered architecture** with domain-oriented handlers:

```
cmd/server/main.go            # Entry point: config → DB → Redis → router
internal/api/
  router.go                   # Route registration, middleware wiring
  handler/                    # HTTP handlers (one per domain module)
  middleware/                 # JWT, CORS, request logging
internal/model/
  base.go                     # All GORM models + BaseModel/BaseModelWithCompany
internal/pkg/
  database/                   # GORM initialization (PostgreSQL)
  redis/                      # go-redis initialization + helpers
  response/                   # Unified API response envelope
  errors/                     # Common errors + BizError wrapper
  billno/                     # Bill number generation (reserved)
  price/                      # Price calculation utilities (reserved)
  stock/                      # Stock calculation utilities (reserved)
internal/repository/          # Reserved for future repository layer
internal/service/             # Reserved for future service layer
internal/scheduler/           # Reserved for future scheduled jobs
```

**Key patterns:**

- **Handler pattern**: Each domain has one handler struct (e.g., `ProductHandler`) that receives `*gorm.DB` via constructor `NewXxxHandler(db)`. Handlers directly query the DB — there is currently no service/repository abstraction layer. `repository/` and `service/` directories are empty and reserved for future refactoring.
- **Multi-tenancy**: All tenant-scoped models embed `BaseModelWithCompany` which adds `CompanyID`. The JWT middleware injects `companyID` into Gin context; handlers filter queries by it.
- **Unified response format** (`internal/pkg/response/`):
  ```json
  { "code": 200, "message": "success", "data": { ... } }
  ```
  Pagination wraps data as `{ list, total, page, pageSize }`.
  - `code === 200` for success
  - `code === 4000-4999` for business errors (validation, duplicate, not found, no permission)
- **JWT**: Access token (2h) + Refresh token (7d). Claims include `userID`, `username`, `roleID`, `deptID`, `companyID`. Middleware injects these into Gin context; retrieve via `middleware.GetUserID(c)`, `GetCompanyID(c)`, etc.
- **Seed data**: On first startup, `handler.InitSeedData(db)` creates a default company, department, role, and admin user (`admin` / `admin123`).
- **Swagger**: Handlers use `swaggo` annotations (`@Summary`, `@Router`, etc.). Run `make swagger` to regenerate `docs/`.

### Frontend (`frontend/`)

```
src/
  api/client.ts               # Axios instance with Bearer token interceptor + 401 redirect
  composables/useCrud.ts      # Reusable CRUD composable (list/create/edit/delete/search/pagination)
  stores/auth.ts              # Pinia auth store (token, user, login/logout)
  router/index.ts             # Vue Router setup
  layouts/MainLayout.vue      # Admin layout with sidebar
  views/                      # Domain views matching backend modules
```

**Key patterns:**

- **CRUD composable**: `useCrud<T>({ baseUrl, defaultForm })` returns reactive state + methods for standard CRUD pages. Most list views use this to avoid boilerplate.
- **API client**: Axios baseURL is empty (relies on Vite proxy or same-origin). Automatically attaches `Authorization: Bearer <token>`.
- **Response handling**: Frontend accepts both `code === 0` and `code === 200` as success (legacy compat).

## Configuration

Backend config is loaded from `configs/config.yaml` and can be overridden by environment variables prefixed with `ZHIZHANG_` (e.g., `ZHIZHANG_APP_PORT=8388`). See `.env.example` for all supported variables.

Key defaults:
- Backend port: `8388`
- PostgreSQL: `localhost:5432/zhizhang`
- Redis: `localhost:6379`

## Development Workflow

1. Start infrastructure: `make infra-up`
2. Run backend: `make run` (or `make dev` with air)
3. In another terminal: `cd frontend && npm run dev`
4. Swagger UI: `http://localhost:8388/swagger/index.html`
5. Default login: `admin` / `admin123`
