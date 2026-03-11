## Task
- Fix backend CORS handling so the web app can call the API from the browser without credentialed cross-origin failures on login and other authenticated requests.

## Plan / TODO
- [x] Reproduce the failure from the browser console details and identify the affected request path and CORS rule.
- [x] Inspect current backend CORS middleware and config defaults.
- [ ] Update CORS behavior to return explicit allowed origins instead of `*` when credentials are involved.
- [ ] Verify preflight and actual auth requests from the web app expectations.
- [ ] Run backend/web verification and summarize any deployment/runtime follow-up needed.

## Notes
- Browser error shows `Cross-Origin Request Blocked` for `http://shadywrldserver:8082/api/auth/login`.
- The browser specifically reports `Credential is not supported if the CORS header 'Access-Control-Allow-Origin' is '*'`.
- Current middleware unconditionally returns `Access-Control-Allow-Origin: *` and `Access-Control-Allow-Credentials: true`, which browsers reject for credentialed requests.
