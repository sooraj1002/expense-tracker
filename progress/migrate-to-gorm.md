# Task: Migrate to GORM

## Problem
Raw SQL database connection is having issues connecting to the correct database, even though the DSN is correct. GORM will handle this better and provide better migration support.

## Solution Plan
Switch from `database/sql` to GORM for all database operations.

## Benefits of GORM
1. Auto-creates database if it doesn't exist (with proper DSN)
2. Built-in migration system using models
3. Cleaner query syntax
4. Better error handling
5. Automatic model scanning and migrations

## Files to Modify
- `db/db.go` - Replace with GORM initialization
- `db/migrate.go` - Replace with GORM AutoMigrate
- `models/*.go` - Add GORM struct tags
- `api/handlers/*.go` - Replace raw SQL queries with GORM queries
- `go.mod` - Add GORM dependencies

## Implementation Steps
- [ ] Add GORM dependencies (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- [ ] Update all model structs with GORM tags
- [ ] Rewrite `db/db.go` to use GORM
- [ ] Replace manual migrations with GORM AutoMigrate
- [ ] Update all handlers to use GORM instead of raw SQL:
  - [ ] auth.go
  - [ ] categories.go
  - [ ] accounts.go
  - [ ] expenses.go
  - [ ] patterns.go
- [ ] Remove old migration SQL files (no longer needed)
- [ ] Test the changes

## GORM Features to Use
- Auto-create database
- AutoMigrate for schema management
- Model associations
- Transactions
- Scopes for reusable queries
