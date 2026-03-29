# Expense Tracker — Go Backend

## Purpose
REST API backend for a personal expense tracker. Stores expenses, accounts, and categories. Designed to be a thin storage layer — heavy classification logic lives in the Android app. Also serves as the target for the voice server's tool calls.

## Stack
- **Go** with Gin framework
- **PostgreSQL** via GORM (ORM)
- **JWT** auth (access + refresh tokens, cookies: `tracker_access`, `tracker_refresh`)
- Runs in Docker, internal port **8080**, exposed as **8082** on the host

## Running
```bash
# Local
go run . serve

# Docker (from expenses-stack/)
docker compose up backend
docker compose build backend && docker compose up -d --force-recreate backend
```

## Key structure
```
cmd/serve.go          — HTTP server setup, graceful shutdown
api/router.go         — All route definitions
api/handlers/         — One file per domain (expenses, accounts, categories, etc.)
api/middleware/       — auth.go (JWT), cors.go, logger.go
models/               — GORM models (Expense, Account, Category, MerchantPattern, etc.)
db/migrate.go         — Auto-migration on startup
config/               — Loads env vars into AppConfig struct
```

## Auth
All `/api/*` routes require `Authorization: Bearer <token>` except:
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/admin/*` (no JWT — be careful)

`middleware.AuthMiddleware()` sets `userID` (uuid.UUID) and `email` (string) in Gin context.

## Expense model (key fields)
```go
Amount      decimal   // required, > 0
CategoryID  uuid      // required — must exist
AccountID   uuid      // required — must exist, balance updated on create/delete
Date        time.Time // required
Description string    // optional
Tags        []string  // defaults to ["misc"]
Verified    bool      // defaults to true on manual create
```
Creating an expense runs a DB transaction: inserts expense + syncs tags + updates account `current_balance` and `total_spent`.

## CORS
Controlled by `ALLOWED_ORIGINS` env var (comma-separated). The voice server calls the backend server-to-server so CORS doesn't apply to those calls.

## Environment variables
```
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
PORT=8080
JWT_SECRET
JWT_EXPIRY=24h
ALLOWED_ORIGINS=http://localhost:3000,...
ENVIRONMENT=development|production
```

## Development process
Before making code changes for any task, create `./progress/{task}.md` with a plan and get it reviewed before writing code. Keep it updated as work progresses.

## Known issues / history
- `middleware/logger.go` had a `fmt.Printf` debug statement inside `sanitizeHeaders` that caused stdout blocking when Docker's log buffer filled up, making all requests hang. Fixed — do not add `fmt.Print*` calls anywhere in middleware.
- The Docker bridge interface (`br-*`) can lose its host IP if NetworkManager interferes. Fixed by `/etc/NetworkManager/conf.d/docker.conf` marking bridge interfaces as unmanaged.
