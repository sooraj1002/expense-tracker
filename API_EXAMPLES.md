# API Endpoint Examples (Backend Canonical)

Base URL: `{{base_url}}`

Use `Authorization: Bearer {{jwt}}` for all endpoints except `/health`, `/api/auth/register`, `/api/auth/login`, and `/api/auth/refresh`.

## Health

```bash
curl {{base_url}}/health
```

## Authentication

- Register
```bash
curl -X POST {{base_url}}/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

- Login
```bash
curl -X POST {{base_url}}/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

- Refresh token
```bash
curl -X POST {{base_url}}/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "token": "{{existing_jwt}}"
  }'
```

- Current user
```bash
curl {{base_url}}/api/auth/me -H "Authorization: Bearer {{jwt}}"
```

- Register device
```bash
curl -X POST {{base_url}}/api/auth/devices/register \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "pixel9_unique_id",
    "deviceName": "Google Pixel 9"
  }'
```

## Categories

- List
```bash
curl {{base_url}}/api/categories -H "Authorization: Bearer {{jwt}}"
```

- Create
```bash
curl -X POST {{base_url}}/api/categories \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Groceries",
    "color": "#4CAF50"
  }'
```

- Update
```bash
curl -X PUT {{base_url}}/api/categories/CATEGORY_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Food & Groceries",
    "color": "#8BC34A"
  }'
```

- Delete
```bash
curl -X DELETE {{base_url}}/api/categories/CATEGORY_ID \
  -H "Authorization: Bearer {{jwt}}"
```

## Accounts

- List
```bash
curl {{base_url}}/api/accounts -H "Authorization: Bearer {{jwt}}"
```

- Create
```bash
curl -X POST {{base_url}}/api/accounts \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "HDFC Bank",
    "initialBalance": 50000.00
  }'
```

- Update
```bash
curl -X PUT {{base_url}}/api/accounts/ACCOUNT_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "HDFC Savings Account",
    "initialBalance": 55000.00
  }'
```

- Delete (fails if expenses exist for the account)
```bash
curl -X DELETE {{base_url}}/api/accounts/ACCOUNT_ID \
  -H "Authorization: Bearer {{jwt}}"
```

- Summary
```bash
curl {{base_url}}/api/accounts/summary -H "Authorization: Bearer {{jwt}}"
```

- Account expenses (filters: `year`, `month`, `page`, `limit` [1-100, default 20])
```bash
curl "{{base_url}}/api/accounts/ACCOUNT_ID/expenses?page=1&limit=20&year=2025&month=11" \
  -H "Authorization: Bearer {{jwt}}"
```

## Expenses

- List (pagination + filters)
  - `page` (default 1)
  - `limit` (default 20, max 100)
  - `year` / `month` (legacy)
  - `period` one of `today|week|month|year`
  - `startDate`, `endDate` (YYYY-MM-DD, inclusive)
  - `accountId`, `categoryId`
  - `tags` (comma-separated; matches any overlap)
  - `sort` one of `date` (default) or `updated`

```bash
curl "{{base_url}}/api/expenses?page=1&limit=20&period=month&accountId=ACCOUNT_ID&categoryId=CATEGORY_ID&tags=food,subway&sort=updated" \
  -H "Authorization: Bearer {{jwt}}"
```

**Response shape**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "userId": "uuid",
      "amount": 1500.0,
      "category": {
        "categoryId": "uuid",
        "categoryName": "Food",
        "color": "#FF5722",
        "isDefault": false
      },
      "accountId": "uuid",
      "date": "2025-11-02T10:30:00Z",
      "description": "Lunch at Subway",
      "tags": ["food", "Subway"],
      "verified": true,
      "createdAt": "timestamp",
      "updatedAt": "timestamp"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "totalCount": 42,
    "totalPages": 3,
    "totalAmount": 52340.75
  }
}
```

- Create (tags default to `["misc"]`; backend marks `verified=true`)
```bash
curl -X POST {{base_url}}/api/expenses \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1500.00,
    "categoryId": "CATEGORY_ID",
    "accountId": "ACCOUNT_ID",
    "date": "2025-11-02T10:30:00Z",
    "description": "Lunch at Subway",
    "tags": ["food", "Subway"]
  }'
```

- Update (any subset of fields)
```bash
curl -X PUT {{base_url}}/api/expenses/EXPENSE_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1650.00,
    "categoryId": "CATEGORY_ID",
    "accountId": "ACCOUNT_ID",
    "date": "2025-11-03T12:00:00Z",
    "description": "Team lunch",
    "tags": ["food", "team"],
    "verified": true
  }'
```

- Delete
```bash
curl -X DELETE {{base_url}}/api/expenses/EXPENSE_ID \
  -H "Authorization: Bearer {{jwt}}"
```

- All expense tags (alphabetical)
```bash
curl {{base_url}}/api/expenses/tags -H "Authorization: Bearer {{jwt}}"
```

## Merchant Patterns

- List (optional `isActive=true|false`)
```bash
curl "{{base_url}}/api/merchant-patterns?isActive=true" \
  -H "Authorization: Bearer {{jwt}}"
```

- Create (match types: `exact`, `contains`)
```bash
curl -X POST {{base_url}}/api/merchant-patterns \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "merchantName": "Swiggy",
    "categoryId": "CATEGORY_ID",
    "matchType": "contains"
  }'
```

- Update
```bash
curl -X PUT {{base_url}}/api/merchant-patterns/PATTERN_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "categoryId": "CATEGORY_ID",
    "matchType": "exact",
    "isActive": true
  }'
```

- Delete
```bash
curl -X DELETE {{base_url}}/api/merchant-patterns/PATTERN_ID \
  -H "Authorization: Bearer {{jwt}}"
```

- Match test
```bash
curl -X POST {{base_url}}/api/merchant-patterns/match \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "merchantName": "SWIGGY BANGALORE"
  }'
```

## Notes

- Pagination defaults: `page=1`, `limit=20` (max 100).
- Expense filters support `period`, `startDate/endDate`, `accountId`, `categoryId`, `tags`, `year/month`, and `sort=date|updated`. There is no server-side text search.
- Expenses return embedded category details and pagination metadata including `totalAmount` for the filtered set.
- Tags are the flexible labeling system (include merchant names as tags if needed).
