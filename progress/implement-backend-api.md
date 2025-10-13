# Task: Implement Backend REST API

## Overview

Implement the full RESTful API backend service based on BACKEND_API.md specification. This will transform the current CLI-based application into a complete API server with PostgreSQL database, JWT authentication, and all required endpoints for the expense tracker Android app.

## Current State

- Basic Cobra CLI setup
- SQLite database initialization (needs migration to PostgreSQL)
- Zap logger initialized
- No HTTP server or API endpoints

## Target State

- Complete REST API server using Gin framework
- PostgreSQL database with proper schema and migrations
- JWT-based authentication system
- All endpoints from BACKEND_API.md implemented
- Proper error handling and validation
- CORS configuration for mobile app
- Structured logging for all operations

## Architecture

```
expense-tracker/
├── main.go                    # Entry point - starts HTTP server
├── cmd/
│   ├── root.go               # CLI commands (if needed)
│   └── server.go             # Server start command
├── api/
│   ├── router.go             # Route definitions
│   ├── middleware/
│   │   ├── auth.go          # JWT authentication middleware
│   │   ├── cors.go          # CORS middleware
│   │   └── logger.go        # Request logging middleware
│   └── handlers/
│       ├── auth.go          # Auth endpoints
│       ├── categories.go    # Category CRUD
│       ├── accounts.go      # Account CRUD
│       ├── expenses.go      # Expense CRUD
│       ├── transactions.go  # Transaction handling
│       ├── merchants.go     # Merchant management
│       ├── patterns.go      # Merchant patterns
│       ├── locations.go     # Location tracking
│       └── sync.go          # Sync operations
├── db/
│   ├── db.go                # Database connection
│   ├── migrations/          # SQL migration files
│   └── postgres.go          # PostgreSQL specific setup
├── models/
│   ├── user.go
│   ├── category.go
│   ├── account.go
│   ├── expense.go
│   ├── transaction.go
│   ├── merchant.go
│   ├── pattern.go
│   ├── location.go
│   └── sync.go
├── services/
│   ├── auth.go              # Authentication logic
│   ├── expense.go           # Expense business logic
│   ├── sync.go              # Sync conflict resolution
│   └── pattern_match.go     # Pattern matching logic
├── utils/
│   ├── jwt.go               # JWT token utilities
│   ├── password.go          # Password hashing
│   └── validator.go         # Input validation
├── config/
│   └── config.go            # Configuration management
└── logger/
    └── log.go               # Logger (already exists)
```

## Todo List

### Phase 1: Project Setup & Dependencies
- [ ] Add required dependencies (Gin, PostgreSQL driver, JWT, bcrypt, validator)
- [ ] Update go.mod with new dependencies
- [ ] Create configuration management (environment variables, config file)
- [ ] Set up PostgreSQL connection (replace SQLite)
- [ ] Create database migration system

### Phase 2: Database Schema & Models
- [ ] Create PostgreSQL migration files for all tables:
  - [ ] users table
  - [ ] categories table (with default system categories)
  - [ ] accounts table
  - [ ] expenses table
  - [ ] transactions table
  - [ ] merchant_info table
  - [ ] merchant_patterns table
  - [ ] locations table
  - [ ] sync_status table
  - [ ] devices table
- [ ] Create Go models for all data structures
- [ ] Add proper indexes for query optimization
- [ ] Add foreign key constraints

### Phase 3: Authentication System
- [ ] Implement password hashing utilities (bcrypt)
- [ ] Implement JWT token generation and validation
- [ ] Create authentication middleware
- [ ] Implement POST /api/auth/register endpoint
- [ ] Implement POST /api/auth/login endpoint
- [ ] Implement POST /api/auth/refresh endpoint
- [ ] Implement GET /api/auth/me endpoint
- [ ] Implement POST /api/auth/devices/register endpoint

### Phase 4: Core API - Categories
- [ ] Implement GET /api/categories
- [ ] Implement POST /api/categories
- [ ] Implement PUT /api/categories/:id
- [ ] Implement DELETE /api/categories/:id
- [ ] Add validation for category operations
- [ ] Seed default system categories in migration

### Phase 5: Core API - Accounts
- [ ] Implement GET /api/accounts
- [ ] Implement POST /api/accounts
- [ ] Implement PUT /api/accounts/:id
- [ ] Implement DELETE /api/accounts/:id
- [ ] Implement GET /api/accounts/summary
- [ ] Implement GET /api/accounts/:id/expenses
- [ ] Add balance calculation logic
- [ ] Add validation for account operations

### Phase 6: Core API - Expenses
- [ ] Implement GET /api/expenses (with pagination and filters)
- [ ] Implement POST /api/expenses
- [ ] Implement POST /api/expenses/batch
- [ ] Implement PUT /api/expenses/:id
- [ ] Implement PUT /api/expenses/:id/verify
- [ ] Implement DELETE /api/expenses/:id
- [ ] Add validation for expense operations
- [ ] Implement account balance update logic

### Phase 7: Core API - Transactions
- [ ] Implement POST /api/transactions
- [ ] Implement POST /api/transactions/batch
- [ ] Implement GET /api/transactions (with filters)
- [ ] Add transaction parsing validation
- [ ] Link transactions to expenses

### Phase 8: Core API - Merchants
- [ ] Implement GET /api/merchants
- [ ] Implement POST /api/merchants
- [ ] Implement GET /api/merchants/:id/expenses
- [ ] Add merchant name normalization logic
- [ ] Update merchant stats when expenses created

### Phase 9: Core API - Merchant Patterns
- [ ] Implement GET /api/merchant-patterns
- [ ] Implement POST /api/merchant-patterns
- [ ] Implement PUT /api/merchant-patterns/:id
- [ ] Implement DELETE /api/merchant-patterns/:id
- [ ] Implement POST /api/merchant-patterns/match
- [ ] Add pattern matching logic (exact, contains)
- [ ] Add pattern usage tracking

### Phase 10: Core API - Locations
- [ ] Implement POST /api/locations
- [ ] Implement GET /api/locations/:id/expenses
- [ ] Add reverse geocoding support (optional)
- [ ] Add location-expense linking

### Phase 11: Sync Operations
- [ ] Implement POST /api/sync/incremental
- [ ] Implement POST /api/sync/batch/expenses
- [ ] Implement GET /api/sync/status/:deviceId
- [ ] Implement POST /api/sync/status
- [ ] Add conflict resolution logic (last-write-wins)
- [ ] Add ID mapping for local to server IDs
- [ ] Add batch processing optimization

### Phase 12: Middleware & Error Handling
- [ ] Create authentication middleware (JWT validation)
- [ ] Create CORS middleware for mobile app
- [ ] Create request logging middleware
- [ ] Implement structured error responses
- [ ] Add request validation middleware
- [ ] Add rate limiting (optional, for production)

### Phase 13: Server Setup & Routing
- [ ] Create Gin router with all routes
- [ ] Group routes by resource
- [ ] Apply authentication middleware to protected routes
- [ ] Set up graceful shutdown
- [ ] Add health check endpoint
- [ ] Update main.go to start HTTP server instead of CLI

### Phase 14: Testing & Documentation
- [ ] Add integration tests for auth endpoints
- [ ] Add integration tests for CRUD endpoints
- [ ] Add unit tests for business logic
- [ ] Test sync conflict resolution
- [ ] Create API testing collection (Postman/Thunder Client)
- [ ] Add README with setup instructions
- [ ] Document environment variables

### Phase 15: Production Readiness
- [ ] Add database connection pooling
- [ ] Add proper logging for all operations
- [ ] Add panic recovery middleware
- [ ] Add request timeouts
- [ ] Create Docker setup (optional)
- [ ] Create systemd service file (optional)
- [ ] Add database backup strategy documentation

## Implementation Strategy

### Step 1: Foundation (Phases 1-2)
Start with project structure, dependencies, and database setup. This provides the foundation for all other work.

### Step 2: Authentication (Phase 3)
Implement authentication early since all other endpoints depend on it.

### Step 3: Core Resources (Phases 4-10)
Implement CRUD operations for all resources in order of dependency:
1. Categories (no dependencies)
2. Accounts (no dependencies)
3. Expenses (depends on categories, accounts)
4. Transactions (depends on expenses)
5. Merchants (depends on expenses)
6. Patterns (depends on categories, merchants)
7. Locations (depends on expenses)

### Step 4: Sync & Polish (Phases 11-15)
Implement sync logic, add middleware, testing, and production readiness features.

## Key Technical Decisions

1. **Web Framework**: Gin (fast, popular, good documentation)
2. **Database**: PostgreSQL (production-ready, better for multi-user)
3. **Authentication**: JWT tokens (stateless, works well for mobile)
4. **Password Hashing**: bcrypt (industry standard)
5. **Validation**: go-playground/validator (comprehensive)
6. **Migrations**: Custom SQL files or golang-migrate
7. **ID Generation**: UUIDs (better for distributed systems)
8. **Timestamps**: Store as UTC, let client handle timezone

## Environment Variables Needed

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=expense_user
DB_PASSWORD=secure_password
DB_NAME=expense_tracker
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info
```

## Database Schema Highlights

- All tables have `id` (UUID), `created_at`, `updated_at`
- User isolation through `user_id` foreign key
- Soft deletes where appropriate
- Indexes on frequently queried columns (user_id, date, category_id)
- Foreign key constraints for referential integrity

## API Response Format

All responses follow consistent format:
```json
{
  "success": true,
  "data": {...},
  "error": null
}
```

Error responses:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "INVALID_INPUT",
    "message": "Amount must be positive"
  }
}
```

## Testing Strategy

1. Unit tests for business logic (pattern matching, sync conflict resolution)
2. Integration tests for API endpoints (using testify)
3. Database tests with test database
4. Mock external dependencies
5. Test coverage target: 70%+

## Timeline Estimate

- Phase 1-2 (Setup): 2-3 hours
- Phase 3 (Auth): 2-3 hours
- Phases 4-10 (CRUD): 8-10 hours (1-1.5 hours per resource)
- Phases 11-12 (Sync & Middleware): 3-4 hours
- Phases 13-15 (Integration & Polish): 3-4 hours

**Total: 18-24 hours of development time**

## Acceptance Criteria

- [ ] All endpoints from BACKEND_API.md are implemented
- [ ] JWT authentication works correctly
- [ ] PostgreSQL database is set up with proper schema
- [ ] Sync operations handle conflicts correctly
- [ ] All CRUD operations have proper validation
- [ ] Error responses are consistent and helpful
- [ ] API can be tested with Postman/curl
- [ ] Server starts successfully and handles requests
- [ ] Logging captures important events
- [ ] Code is organized and maintainable

## Dependencies to Add

```
go get -u github.com/gin-gonic/gin
go get -u github.com/lib/pq
go get -u github.com/golang-jwt/jwt/v5
go get -u golang.org/x/crypto/bcrypt
go get -u github.com/go-playground/validator/v10
go get -u github.com/google/uuid
go get -u github.com/joho/godotenv
```

## Notes

- Start with a minimal working version of each endpoint, then enhance
- Use transactions for operations that modify multiple tables
- Keep business logic in services, not handlers
- Validate all user input
- Log all errors with context
- Use prepared statements to prevent SQL injection
- Return appropriate HTTP status codes
- Keep handlers thin - delegate to services

## Next Steps

After this plan is reviewed and approved:
1. Start with Phase 1 - install dependencies and set up configuration
2. Work through phases sequentially
3. Test each endpoint as it's implemented
4. Update this document with progress checkmarks
5. Document any issues or changes to the plan
