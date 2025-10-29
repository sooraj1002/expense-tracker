# GORM + Fiber Migration Plan

## Overview
Migrate the expense-tracker backend from:
- **Web Framework**: Gin → Fiber (v2)
- **Database Layer**: database/sql + lib/pq → GORM (with PostgreSQL driver)

### Goals
- ✅ Maintain 100% feature parity (all 25+ API endpoints)
- ✅ Improve code maintainability with GORM's ORM features
- ✅ Keep all existing business logic intact
- ✅ Maintain authentication & authorization
- ✅ Preserve transaction handling for expense operations
- ✅ Keep existing configuration and utilities

---

## Migration Todo List

### Phase 1: Setup & Dependencies
- [ ] Update go.mod with GORM and Fiber dependencies
  - [ ] Add `gorm.io/gorm`
  - [ ] Add `gorm.io/driver/postgres`
  - [ ] Add `github.com/gofiber/fiber/v2`
  - [ ] Remove `github.com/gin-gonic/gin`
  - [ ] Keep other dependencies (zap, jwt, bcrypt, etc.)
- [ ] Run `go mod tidy` and verify no conflicts

### Phase 2: Models Layer (Add GORM Tags)
- [ ] Update `/models/user.go` with GORM tags
  - [ ] Add UUID type, primary key tag
  - [ ] Add unique index on email
  - [ ] Add timestamps (CreatedAt, UpdatedAt)
- [ ] Update `/models/category.go` with GORM tags
  - [ ] Foreign key to User (nullable for system categories)
  - [ ] Indexes on user_id, is_default
- [ ] Update `/models/account.go` with GORM tags
  - [ ] Foreign key to User
  - [ ] Decimal type for balances
  - [ ] Timestamps
- [ ] Update `/models/expense.go` with GORM tags
  - [ ] Foreign keys to User, Category, Account, Merchant, Location
  - [ ] Multiple indexes (user, category, account, date)
  - [ ] Nullable foreign keys (merchant_id, location_id)
- [ ] Update `/models/transaction.go` with GORM tags
  - [ ] Foreign key to User and Expense
  - [ ] Indexes on processed, timestamp
- [ ] Update `/models/merchant.go` with GORM tags
  - [ ] Unique constraint on (user_id, name)
  - [ ] Array field for aliases (use pq.StringArray)
- [ ] Update `/models/pattern.go` with GORM tags
  - [ ] Unique constraint on (user_id, merchant_name)
  - [ ] Indexes on is_active
- [ ] Update `/models/location.go` with GORM tags
  - [ ] Decimal types for lat/long
  - [ ] Index on timestamp
- [ ] Update `/models/device.go` with GORM tags
  - [ ] Unique constraint on device_id
  - [ ] Timestamps
- [ ] Update `/models/sync.go` with GORM tags
  - [ ] Foreign key to Device
  - [ ] Unique constraint on (user_id, device_id)

### Phase 3: Database Layer Rewrite
- [ ] Rewrite `/db/db.go` to use GORM
  - [ ] Replace `sql.Open()` with `gorm.Open()`
  - [ ] Configure connection pool settings
  - [ ] Keep global DB variable (change type to *gorm.DB)
  - [ ] Update InitDB() function signature
  - [ ] Test connection with `db.Exec("SELECT 1")`
- [ ] Update `/db/migrate.go` migration system
  - [ ] Option A: Keep SQL-based migrations (use GORM's raw SQL)
  - [ ] Option B: Use GORM's AutoMigrate (generates schema from models)
  - [ ] **Decision needed**: Which approach to use?
  - [ ] Maintain schema_migrations table for tracking
- [ ] Update `/cmd/serve.go` to use new GORM DB
  - [ ] Update InitDB call
  - [ ] Update migration call if changed

### Phase 4: Middleware Migration (Gin → Fiber)
- [ ] Rewrite `/api/middleware/cors.go` for Fiber
  - [ ] Use Fiber's CORS middleware or custom implementation
  - [ ] Maintain same CORS policy (allow all origins, credentials)
- [ ] Rewrite `/api/middleware/auth.go` for Fiber
  - [ ] Replace `gin.Context` with `fiber.Ctx`
  - [ ] Use `c.Get("Authorization")` instead of `c.GetHeader()`
  - [ ] Store user data in Fiber locals: `c.Locals("userID", userID)`
  - [ ] Update GetUserID() helper for Fiber context
- [ ] Rewrite `/api/middleware/logger.go` for Fiber
  - [ ] Use `c.Method()`, `c.Path()`, `c.IP()` instead of Gin equivalents
  - [ ] Maintain same Zap logging format

### Phase 5: Handlers Migration (25 functions)
Each handler needs:
- Replace `gin.Context` → `fiber.Ctx`
- Replace `c.ShouldBindJSON()` → `c.BodyParser()`
- Replace `c.JSON(status, data)` → `c.Status(status).JSON(data)`
- Replace `c.Param()` → `c.Params()`
- Replace database/sql queries → GORM queries

#### `/api/handlers/auth.go` (5 functions)
- [ ] Migrate `Register()` handler
  - [ ] Request binding with BodyParser
  - [ ] GORM: `db.Create(&user)` instead of INSERT
  - [ ] Password hashing (keep same)
  - [ ] JWT generation (keep same)
  - [ ] Response with Fiber
- [ ] Migrate `Login()` handler
  - [ ] GORM: `db.Where("email = ?", email).First(&user)`
  - [ ] Password verification (keep same)
  - [ ] Update last_login_at: `db.Model(&user).Update("last_login_at", time.Now())`
  - [ ] JWT generation
- [ ] Migrate `RefreshToken()` handler
  - [ ] Token validation (keep same)
  - [ ] Generate new token
- [ ] Migrate `GetMe()` handler
  - [ ] Get userID from Fiber locals
  - [ ] GORM: `db.First(&user, userID)`
- [ ] Migrate `RegisterDevice()` handler
  - [ ] GORM: FirstOrCreate or Create with OnConflict

#### `/api/handlers/accounts.go` (6 functions)
- [ ] Migrate `GetAccounts()` handler
  - [ ] GORM query: `db.Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&accounts)`
  - [ ] Count: `db.Model(&Account{}).Where("user_id = ?", userID).Count(&total)`
- [ ] Migrate `CreateAccount()` handler
  - [ ] GORM: `db.Create(&account)`
- [ ] Migrate `UpdateAccount()` handler
  - [ ] GORM: `db.Model(&account).Updates(map[string]interface{}{...})`
- [ ] Migrate `DeleteAccount()` handler
  - [ ] Check for expenses: `db.Model(&Expense{}).Where("account_id = ?", id).Count(&count)`
  - [ ] GORM: `db.Delete(&account)`
- [ ] Migrate `GetAccountSummary()` handler
  - [ ] GORM aggregation: `db.Model(&Account{}).Select("SUM(current_balance)").Where("user_id = ?", userID).Scan(&sum)`
- [ ] Migrate `GetAccountExpenses()` handler
  - [ ] Complex query with date filters
  - [ ] GORM query builder with Where clauses
  - [ ] Pagination

#### `/api/handlers/categories.go` (4 functions)
- [ ] Migrate `GetCategories()` handler
  - [ ] GORM: Join or two queries (user categories + default categories)
  - [ ] `db.Where("user_id = ? OR is_default = true", userID).Find(&categories)`
- [ ] Migrate `CreateCategory()` handler
  - [ ] GORM: `db.Create(&category)`
- [ ] Migrate `UpdateCategory()` handler
  - [ ] Verify ownership and not default
  - [ ] GORM: `db.Model(&category).Updates(updates)`
- [ ] Migrate `DeleteCategory()` handler
  - [ ] Check for expense references
  - [ ] GORM: `db.Delete(&category)`

#### `/api/handlers/expenses.go` (4 functions)
- [ ] Migrate `GetExpenses()` handler
  - [ ] Complex dynamic query with filters (month, year, account_id)
  - [ ] GORM query builder with conditional Where
  - [ ] Order by date descending
  - [ ] Pagination
- [ ] Migrate `CreateExpense()` handler **[CRITICAL - Uses Transaction]**
  - [ ] GORM transaction: `db.Transaction(func(tx *gorm.DB) error { ... })`
  - [ ] Create expense: `tx.Create(&expense)`
  - [ ] Update account balance: `tx.Model(&account).Updates(map[string]interface{}{...})`
  - [ ] Rollback on error
- [ ] Migrate `UpdateExpense()` handler
  - [ ] Dynamic updates based on request fields
  - [ ] GORM: `db.Model(&expense).Updates(updates)`
- [ ] Migrate `DeleteExpense()` handler **[CRITICAL - Uses Transaction]**
  - [ ] GORM transaction
  - [ ] Delete expense: `tx.Delete(&expense)`
  - [ ] Reverse account balance update
  - [ ] Rollback on error

#### `/api/handlers/patterns.go` (5 functions)
- [ ] Migrate `GetMerchantPatterns()` handler
  - [ ] GORM: `db.Where("user_id = ?", userID).Find(&patterns)`
  - [ ] Optional filter: `.Where("is_active = ?", true)`
- [ ] Migrate `CreateMerchantPattern()` handler
  - [ ] Check for duplicate: `db.Where("user_id = ? AND merchant_name = ?", ...).First(&existing)`
  - [ ] GORM: `db.Create(&pattern)`
- [ ] Migrate `UpdateMerchantPattern()` handler
  - [ ] Dynamic updates
  - [ ] GORM: `db.Model(&pattern).Updates(updates)`
- [ ] Migrate `DeleteMerchantPattern()` handler
  - [ ] GORM: `db.Delete(&pattern)`
- [ ] Migrate `MatchMerchantPattern()` handler
  - [ ] Query active patterns
  - [ ] Keep matching logic in Go code (not in DB)

### Phase 6: Router Migration
- [ ] Rewrite `/api/router.go` to use Fiber
  - [ ] Create Fiber app: `fiber.New(fiber.Config{...})`
  - [ ] Apply global middleware (recovery, logger, CORS)
  - [ ] Define all routes with Fiber syntax
  - [ ] Public routes: `app.Post("/api/auth/register", handlers.Register)`
  - [ ] Protected route group: `api := app.Group("/api", middleware.AuthMiddleware())`
  - [ ] All 25+ routes defined
  - [ ] Health check endpoint

### Phase 7: Server Startup
- [ ] Update `/cmd/serve.go` for Fiber
  - [ ] Replace http.Server setup with Fiber
  - [ ] Use `app.Listen(":3000")`
  - [ ] Graceful shutdown with `app.ShutdownWithTimeout()`
  - [ ] Keep signal handling (SIGINT, SIGTERM)

### Phase 8: Testing & Validation
- [ ] Test public endpoints
  - [ ] POST /health
  - [ ] POST /api/auth/register
  - [ ] POST /api/auth/login
  - [ ] POST /api/auth/refresh
- [ ] Test authentication
  - [ ] Auth middleware with valid token
  - [ ] Auth middleware with invalid token
  - [ ] Auth middleware with expired token
- [ ] Test user endpoints
  - [ ] GET /api/auth/me
  - [ ] POST /api/auth/devices/register
- [ ] Test account endpoints (all 6)
  - [ ] CRUD operations
  - [ ] Summary and expenses queries
- [ ] Test category endpoints (all 4)
  - [ ] List, create, update, delete
- [ ] Test expense endpoints (all 4)
  - [ ] **Critical**: Test transaction rollback on create failure
  - [ ] **Critical**: Test transaction rollback on delete failure
  - [ ] Test complex filters (month, year, account)
- [ ] Test merchant pattern endpoints (all 5)
  - [ ] Pattern matching logic
- [ ] Test pagination on all list endpoints
- [ ] Test error handling
  - [ ] 400 Bad Request for validation errors
  - [ ] 401 Unauthorized for missing/invalid auth
  - [ ] 403 Forbidden for unauthorized access
  - [ ] 404 Not Found for missing resources
  - [ ] 500 Internal Server Error for DB errors

### Phase 9: Cleanup
- [ ] Remove old Gin imports
- [ ] Remove unused database/sql code
- [ ] Update comments and documentation
- [ ] Run `go mod tidy`
- [ ] Update README if exists

---

## Implementation Strategy

### Approach: Bottom-Up Migration
1. **Models First**: Add GORM tags without breaking existing code
2. **Database Layer**: Replace connection management
3. **Middleware**: Independent components, can be done in parallel
4. **Handlers**: One file at a time, test each before moving on
5. **Router**: Wire everything together
6. **Server**: Final integration

### Critical Decisions Needed

#### Decision 1: Migration Strategy
**Options:**
- **A) Keep SQL migrations**: Use GORM's `db.Exec()` to run existing .sql files
  - ✅ Pros: Explicit control, can review exact schema
  - ❌ Cons: Need to maintain SQL files separately

- **B) Use GORM AutoMigrate**: Let GORM generate schema from models
  - ✅ Pros: Schema and models stay in sync automatically
  - ❌ Cons: Less control over exact DDL, existing migrations need conversion

**Recommendation**: Option A (keep SQL) for production stability, can switch to AutoMigrate later

#### Decision 2: Database Connection Issue
**Current Issue**: Two PostgreSQL instances running (system + docker)
- App connects to system Postgres, not docker container
- Need to either:
  - Stop system Postgres and use docker only
  - Change docker port mapping (e.g., 5433:5432)
  - Fix connection string to use docker IP

**Recommendation**: Stop system postgres or change docker port to avoid confusion

---

## File-by-File Change Summary

### Files to Modify (20 files)

| File | Changes | Complexity |
|------|---------|------------|
| `go.mod` | Add GORM/Fiber, remove Gin | Simple |
| `models/*.go` (10 files) | Add GORM struct tags | Medium |
| `db/db.go` | Replace sql.DB with gorm.DB | Medium |
| `db/migrate.go` | Update for GORM (if needed) | Medium |
| `api/middleware/cors.go` | Rewrite for Fiber | Simple |
| `api/middleware/auth.go` | Replace Context type | Simple |
| `api/middleware/logger.go` | Replace Context type | Simple |
| `api/handlers/auth.go` | Rewrite 5 handlers | High |
| `api/handlers/accounts.go` | Rewrite 6 handlers | High |
| `api/handlers/categories.go` | Rewrite 4 handlers | Medium |
| `api/handlers/expenses.go` | Rewrite 4 handlers (with transactions) | **Critical** |
| `api/handlers/patterns.go` | Rewrite 5 handlers | Medium |
| `api/router.go` | Rewrite router setup | Medium |
| `cmd/serve.go` | Update server startup | Simple |

### Files to Keep Unchanged (7 files)
- `config/config.go` - No changes needed
- `utils/jwt.go` - Framework-agnostic
- `utils/password.go` - Framework-agnostic
- `logger/log.go` - Framework-agnostic
- `main.go` - Just calls cmd.Execute()
- `cmd/root.go` - Just Cobra setup
- `.env` - Configuration file

---

## Testing Plan

### Test Environment Setup
1. Clean database state
2. Run migrations
3. Start server
4. Run test suite

### Test Cases (Grouped by Priority)

#### P0: Critical Path (Must Work)
- [ ] User registration and login
- [ ] JWT authentication
- [ ] Create expense with account balance update (transaction)
- [ ] Delete expense with account balance reversal (transaction)

#### P1: Core Features
- [ ] All account CRUD operations
- [ ] All category CRUD operations
- [ ] All expense CRUD operations
- [ ] Merchant pattern matching

#### P2: Additional Features
- [ ] Device registration
- [ ] Token refresh
- [ ] Pagination on all list endpoints
- [ ] Complex filters (date, account)

#### P3: Edge Cases
- [ ] Authorization checks (can't access other user's data)
- [ ] Validation errors (400 responses)
- [ ] Not found errors (404 responses)
- [ ] Database constraints (unique, foreign key)

### Test Tools
- Manual testing with `curl` or Postman
- Optional: Write integration tests with Go's testing package
- Optional: Use httpie for cleaner CLI testing

---

## Rollback Strategy

### Git Workflow
1. Create feature branch: `git checkout -b feat/gorm-fiber-migration`
2. Commit after each phase
3. Tag current main as `pre-gorm-fiber` before merging
4. If issues found, can revert to tag

### Database Rollback
- Keep SQL migrations to allow downgrade
- Don't delete old migration files
- Schema remains compatible

### Deployment Strategy
- Test locally first
- Deploy to staging environment
- Run full test suite
- Monitor for errors
- If stable, deploy to production

---

## Timeline Estimate

| Phase | Estimated Time | Complexity |
|-------|---------------|------------|
| Phase 1: Dependencies | 15 min | Low |
| Phase 2: Models | 1-2 hours | Medium |
| Phase 3: Database Layer | 1 hour | Medium |
| Phase 4: Middleware | 1 hour | Low |
| Phase 5: Handlers | 4-6 hours | High |
| Phase 6: Router | 1 hour | Medium |
| Phase 7: Server | 30 min | Low |
| Phase 8: Testing | 2-3 hours | High |
| Phase 9: Cleanup | 30 min | Low |
| **Total** | **11-15 hours** | - |

---

## Risk Assessment

### High Risk Items
1. **Transaction handling in expenses** - If wrong, could corrupt account balances
2. **Authentication middleware** - If broken, security vulnerability
3. **Data type conversions** - Money fields must maintain precision

### Medium Risk Items
1. **Complex queries** (GetExpenses, GetAccountExpenses) - Could have performance issues
2. **Pagination logic** - Off-by-one errors common
3. **Migration system** - Schema drift if not careful

### Low Risk Items
1. **Simple CRUD handlers** - Straightforward GORM operations
2. **Configuration** - No changes needed
3. **Utilities** - Already framework-agnostic

---

## Success Criteria

✅ **Migration is successful when:**
1. All 25+ API endpoints return correct responses
2. All tests pass (manual or automated)
3. No data corruption in database
4. Transactions work correctly (rollback on error)
5. Authentication and authorization work
6. No performance regression
7. Code is cleaner and more maintainable
8. Server starts without errors
9. Graceful shutdown works

---

## Notes & Open Questions

### Questions for Review
1. **Migration strategy**: Keep SQL migrations or use AutoMigrate?
2. **Database connection**: Fix system vs docker postgres conflict first?
3. **Testing approach**: Manual testing or write test suite?
4. **Deployment**: Any specific deployment requirements?
5. **Error handling**: Keep current error response format or change?

### Additional Considerations
- GORM logging: Enable for development, disable for production
- Connection pool tuning: May need adjustment after GORM
- Query performance: Monitor for N+1 query problems
- Soft deletes: Consider using GORM's soft delete feature
- Preloading: Use GORM's Preload for relations if needed

---

## Status: AWAITING APPROVAL

**Next Steps:**
1. Review this plan
2. Answer open questions
3. Approve to proceed with implementation
4. Start with Phase 1 (dependencies)

**Last Updated**: 2025-10-28
**Created By**: Claude Code
