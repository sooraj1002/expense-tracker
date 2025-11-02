# Task: Auto-create Database if it Doesn't Exist

## Problem
The application tries to connect to `expense_tracker` database, but when it doesn't exist, it falls back to the default `postgres` database silently. This causes data to be stored in the wrong database.

## Solution Plan
1. Modify `db.InitDB()` to accept both the target DSN and database name
2. Before connecting to the target database:
   - Connect to the default `postgres` database first
   - Check if the target database exists
   - If not, create it
   - Close the connection to `postgres`
3. Then connect to the target database as usual

## Files to Modify
- `db/db.go` - Add database creation logic
- `cmd/serve.go` - Pass database name to InitDB function

## Implementation Steps
- [x] Create a new function `ensureDatabaseExists()` in `db/db.go`
- [x] This function will:
  - [x] Connect to `postgres` database
  - [x] Query to check if target database exists
  - [x] Create database if it doesn't exist
  - [x] Close the connection
- [x] Modify `InitDB()` to call `ensureDatabaseExists()` before connecting
- [x] Update `cmd/serve.go` to pass the database config parameters
- [x] Ensure all code uses global `db.DB` variable:
  - [x] Modified `RunMigrations()` to use global `db.DB` instead of parameter
  - [x] Verified all handlers already use `db.DB`
- [ ] Test the changes

## Testing
- Run the server and verify it creates the database if missing
- Verify the correct database is being used (check the log output)
