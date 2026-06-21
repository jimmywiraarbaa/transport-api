# Transport API

Backend REST API for **Transport Care** — a vehicle maintenance management system that tracks part replacement schedules based on kilometer intervals and time periods.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| Framework | Gin |
| Database | PostgreSQL 16 (pgx/v5 connection pool) |
| Query Layer | sqlc (type-safe SQL code generation) |
| Migrations | golang-migrate |
| Auth | JWT (access + refresh tokens) |
| Config | Viper |
| Hot Reload | Air |

## Architecture

Clean Architecture pattern (bxcodec/go-clean-arch v4) — each feature is a self-contained module with 4 layers:

```
transport-api/
  cmd/api/              entrypoint — boots config, DB pool, container, routes
  app/                  DI container + route wiring
  internal/             shared infrastructure
    config/             Viper config loader
    database/           pgxpool + sqlc generated code
    middleware/         Recovery, Logger, CORS, RequireAuth
    utils/jwt/          JWT token manager
  migrations/           SQL schema + seed data
  queries/              sqlc query files (*.sql)
  sqlc.yaml             sqlc configuration
  auth/                 domain → repository → usecase → delivery
  vehicletype/          domain → repository → usecase → delivery
  maintenancepart/      domain → repository → usecase → delivery
  schedulerule/         domain → repository → usecase → delivery
  vehicle/              domain → repository → usecase → delivery
  maintenancerecord/    domain → repository → usecase → delivery
  maintenancealert/     domain → repository → usecase → delivery (core read-model)
```

### Layer responsibilities

| Layer | Responsibility |
|---|---|
| `domain/` | Entity structs, request/response DTOs, status constants |
| `repository/` | SQL queries via sqlc, DB ↔ domain mapping |
| `usecase/` | Business logic (alert computation, odometer validation, etc.) |
| `delivery/` | HTTP handlers, Gin route registration, JSON serialization |

## Database Schema

```
users                Auth users (email, username, password_hash)
vehicle_types        Master: mobil, motor, truk
vehicles             User vehicles (plate, brand, model, odometer, initial_odometer_km)
maintenance_parts    Master: oli-mesin, filter-oli, kampas-rem, etc.
schedule_rules       Maintenance intervals per part + vehicle type
maintenance_records  Logged service events (odometer, date, cost, technician)
```

ER relationships:

```
users 1───* vehicles *───1 vehicle_types
vehicles 1───* maintenance_records *───1 maintenance_parts
maintenance_parts 1───* schedule_rules *───1 vehicle_types
```

## API Endpoints

### Auth (public)

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login, returns access + refresh tokens |
| POST | `/api/v1/auth/refresh` | Exchange refresh token for new access token |

### Auth (protected)

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/auth/me` | Get current user profile |

### Master Data

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/vehicle-types` | List all vehicle types |
| GET | `/api/v1/maintenance-parts` | List all maintenance parts |
| GET | `/api/v1/schedule-rules` | List all schedule rules |

### Vehicles

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/vehicles` | List user's vehicles |
| POST | `/api/v1/vehicles` | Create vehicle |
| GET | `/api/v1/vehicles/:id` | Get vehicle detail |
| PUT | `/api/v1/vehicles/:id` | Update vehicle |
| DELETE | `/api/v1/vehicles/:id` | Delete vehicle |
| PATCH | `/api/v1/vehicles/:id/odometer` | Update current odometer |

### Maintenance

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/vehicles/:id/alerts` | **Compute maintenance alert status per part** |
| GET | `/api/v1/vehicles/:id/maintenance-records` | List records for a vehicle |
| POST | `/api/v1/vehicles/:id/maintenance-records` | Create record (auto-advances odometer) |
| GET | `/api/v1/maintenance-records/:id` | Get record detail |
| PUT | `/api/v1/maintenance-records/:id` | Update record |
| DELETE | `/api/v1/maintenance-records/:id` | Delete record |

### Other

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Liveness probe (public) |

## Alert Status Logic

The core feature — `/api/v1/vehicles/:id/alerts` computes per-part maintenance status:

| Status | Meaning |
|---|---|
| `ok` | Comfortably within thresholds |
| `due_soon` | Within 500 km or 14 days of threshold |
| `overdue` | Threshold passed (km or date exceeded) |

**Baseline:** Uses `initial_odometer_km` (captured at vehicle creation) and vehicle creation date as baseline when no prior maintenance record exists.

**Trigger modes:**

| Mode | Behavior |
|---|---|
| `or` | Whichever threshold is more urgent |
| `and` | Both thresholds must trigger |
| `km_only` | Only kilometer interval |
| `date_only` | Only time interval |

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 16+ (or use Docker)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [sqlc](https://sqlc.dev) v1.31+
- [Air](https://github.com/air-verse/air) for hot reload

### Docker (recommended)

```bash
cd transport-api
cp .env.example .env
docker compose up -d --build
```

This starts:
- PostgreSQL on `localhost:5434`
- API on `localhost:8080`

Migrations run automatically on container startup.

### Local Development

```bash
cd transport-api
cp .env.example .env

# 1. Start PostgreSQL
make db-up              # or: docker compose up -d postgres

# 2. Apply migrations
make migrate-up

# 3. Run with hot reload
make dev
```

### Default Credentials

The seed migration creates default master data. Register a new user via `/api/v1/auth/register` or use:
- Email: `admin@transport.test` / Password: `admin123`

## Configuration

Environment variables (see `.env.example`):

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Environment |
| `APP_PORT` | `8080` | Server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5434` | PostgreSQL port |
| `DB_USER` | `transport` | DB user |
| `DB_PASSWORD` | `transport_secret` | DB password |
| `DB_NAME` | `transport_db` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `DB_MAX_CONNS` | `20` | Connection pool size |
| `JWT_ACCESS_SECRET` | — | JWT access token secret |
| `JWT_REFRESH_SECRET` | — | JWT refresh token secret |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime (7 days) |

## Makefile Commands

```bash
make help             # list all commands
make dev              # run with Air hot reload
make build            # build binary to ./tmp/api
make run              # build then run
make test             # run tests with coverage
make tidy             # go mod tidy
make lint             # run golangci-lint
make migrate-up       # apply all migrations
make migrate-down     # rollback last migration
make migrate-create   # create migration pair: make migrate-create name=add_column
make sqlc-gen         # regenerate sqlc code
make docker-up        # start full stack via docker compose
make docker-down      # stop full stack
```

## sqlc Code Generation

SQL queries live in `queries/*.sql`. After modifying queries or migrations:

```bash
make sqlc-gen         # or: sqlc generate
```

Generated code goes to `internal/database/sqlcgen/`. **Never edit generated files manually.**

### sqlc config highlights

- Engine: PostgreSQL with `pgx/v5` driver
- UUID mapped to `github.com/google/uuid.UUID`
- Timestamps mapped to `time.Time`
- Numeric mapped to `github.com/shopspring/decimal.Decimal`
- Null types emitted as pointers

## Migrations

Managed by golang-migrate. Migration files in `migrations/`:

| File | Description |
|---|---|
| `000001_init_schema` | Tables, enums, indexes, triggers |
| `000002_seed_masters` | Seed vehicle types, parts, schedule rules |
| `000003_add_username` | Add username column to users |
| `000004_add_initial_odometer` | Add initial_odometer_km to vehicles |
