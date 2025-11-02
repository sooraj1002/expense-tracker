# Task: Fix Server Startup Issues

## Problems Found

1. **Compilation Error**: Unused import `uuid` in `api/handlers/auth.go`
2. **Database Connection Issue**: Server connecting to `postgres` database instead of `expense_tracker`
3. **Migration Conflicts**: Old schema in `postgres` database causing constraint name mismatches

## Root Cause

The empty database password in `.env` was causing PostgreSQL's libpq to ignore the `dbname` parameter in the DSN, resulting in connections falling back to the default `postgres` database.

## Solutions Implemented

### 1. Fixed Unused Import
- [x] Removed unused `github.com/google/uuid` import from `api/handlers/auth.go:9`

### 2. Fixed Database Connection
- [x] Set password for postgres user in Docker container: `ALTER USER postgres WITH PASSWORD 'postgres'`
- [x] Updated `.env` file to use password: `DB_PASSWORD=postgres`

### 3. Cleaned Up Wrong Database
- [x] Dropped old tables from `postgres` database that were created during failed migrations
- [x] Verified tables are now created in correct `expense_tracker` database

## Verification

- [x] Server starts without errors
- [x] Migrations complete successfully
- [x] All 8 tables created in `expense_tracker` database
- [x] `/health` endpoint responds successfully
- [x] User registration endpoint works correctly
- [x] Data is stored in the correct database

## Infrastructure Documentation Added

- [x] Added database information to `CLAUDE.md`:
  - Database: PostgreSQL running in Docker container
  - Container name: `postgres15-dev`
  - Database name: `expense_tracker`
  - Default user: `postgres`
  - Password: `postgres`

## Status

✅ **COMPLETED** - Server is now running successfully on port 3000
