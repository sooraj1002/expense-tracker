## Task
- Add monthly per-category budgets to the backend and Android app.
- Let the user set or clear a budget for any category, including system defaults.
- Show current-month category spend and clearly flag over-budget categories on the Android home screen.

## Plan / TODO
- [x] Add a user-scoped category budget model and include it in GORM AutoMigrate.
- [x] Enrich category responses with the authenticated user's monthly budget and authoritative current-month spend.
- [x] Add authenticated endpoints to set or clear a category budget, validating category visibility and positive amounts.
- [x] Extend the Android API/types/store to persist budget values returned with categories.
- [x] Add budget controls to the Android Categories screen without unlocking edits to system category names/colors.
- [x] Add a current-month budget section to the Android home screen and refresh it whenever the home screen regains focus.
- [x] Run Go formatting/tests and Android lint/bundle verification; document baseline toolchain issues.

## Design review
- Budgets belong in a separate `(user_id, category_id)` table because default categories are shared globally while budget amounts are personal.
- A budget is monthly and recurring; `null` clears it. The initial feature intentionally has no rollover or historical schedule.
- Expense currency is respected: a category budget is treated as INR and only INR expenses count toward it. This avoids silently combining unlike currencies.
- Existing category create/update payloads stay compatible. Budget updates use a focused endpoint so default categories can receive budgets while remaining otherwise immutable.

## Notes
- The web client is explicitly out of scope for this task.
- Backend verification: `go test ./...` passes in the Go 1.24 container; modified Go files were formatted and pass `git diff --check`.
- Android verification: the touched screens pass ESLint with warnings only, and a production Android Metro bundle completes successfully.
- Full Android `tsc` remains blocked by existing unrelated project errors (custom icon/color-picker typings and older screens/services). Jest also fails before test execution because its current config does not transform React Redux's ESM bundle.
- A Gradle APK build could not start because this host does not have Java/`JAVA_HOME`. The Compose backend image build also hit Docker DNS while the Go 1.22 base image tried to download the Go 1.24.3 toolchain; direct Go 1.24 container tests passed.
