# Fix Expense Update Date Handling

## Todo
- [x] Inspect `UpdateExpense` handler to confirm date/account fields are patched.
- [x] Ensure update payload persists `date` changes (and other optional fields) when editing.
- [x] Run `go test ./...` to confirm no regressions.
- [ ] Provide reviewer notes summarising the backend tweak.

## Notes
- Added support for updating `account_id` and `date` within the dynamic updates map so edits are persisted.
- Reviewer pointers: change is confined to `api/handlers/expenses.go` update map; gofmt applied; tests run with `go test ./...`.
