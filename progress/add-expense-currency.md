## Task
- Add a `currency` column to the `expenses` table. Default `INR`, must be one of `USD`, `SGD`, `INR`.

## Plan / TODO
- [x] Add `Currency` field to `models.Expense` (gorm: varchar(3), not null, default 'INR', check constraint restricting to USD/SGD/INR) — GORM AutoMigrate is the actual source of truth for schema (confirmed `db/migrate.go` only calls `AutoMigrate`; the numbered SQL files under `db/migrations/` are stale/unused, e.g. `expenses.source` column in `004_create_expenses_table.sql` has no corresponding model field).
- [x] Add `Currency` (optional, `oneof=USD SGD INR`) to `CreateExpenseRequest`, default to `INR` server-side when omitted.
- [x] Add `Currency` (optional pointer, `oneof=USD SGD INR`) to `UpdateExpenseRequest`.
- [x] Add `Currency` to `ExpenseResponse`.
- [x] Wire `Currency` through `api/handlers/expenses.go`: `CreateExpense`, `UpdateExpense`, `GetExpenses` response mapping, and CSV export.
- [x] Build to confirm it compiles (verified via `golang:1.24` container with `--network host`; local host has no `go` binary and default docker bridge network has no DNS/internet).

## Notes
- No separate currency table/enum exists; using a CHECK constraint + Go `oneof` binding tag, consistent with existing `source` check-constraint precedent (though currently unused) and the app's use of Go-level validation elsewhere.
