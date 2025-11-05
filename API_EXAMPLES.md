# API Endpoint Examples

Base URL: `{{base_url}}`

## Table of Contents
- [Health Check](#health-check)
- [Authentication](#authentication)
- [Categories](#categories)
- [Accounts](#accounts)
- [Expenses](#expenses)
- [Merchant Patterns](#merchant-patterns)

---

## Health Check

### Check API Health
```bash
curl {{base_url}}/health
```

---

## Authentication

### Register New User
```bash
curl -X POST {{base_url}}/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "createdAt": "timestamp",
      "updatedAt": "timestamp"
    },
    "token": "jwt_token_here"
  }
}
```

### Login
```bash
curl -X POST {{base_url}}/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Refresh Token
```bash
curl -X POST {{base_url}}/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_existing_jwt_token"
  }'
```

### Get Current User Profile
```bash
curl {{base_url}}/api/auth/me \
  -H "Authorization: Bearer {{jwt}}"
```

### Register Device
```bash
curl -X POST {{base_url}}/api/auth/devices/register \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "pixel9_unique_id",
    "deviceName": "Google Pixel 9"
  }'
```

---

## Categories

### Get All Categories
```bash
curl {{base_url}}/api/categories \
  -H "Authorization: Bearer {{jwt}}"
```

### Create Category
```bash
curl -X POST {{base_url}}/api/categories \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Groceries",
    "color": "#4CAF50"
  }'
```

### Update Category
```bash
curl -X PUT {{base_url}}/api/categories/CATEGORY_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Food & Groceries",
    "color": "#8BC34A"
  }'
```

### Delete Category
```bash
curl -X DELETE {{base_url}}/api/categories/CATEGORY_ID \
  -H "Authorization: Bearer {{jwt}}"
```

---

## Accounts

### Get All Accounts
```bash
curl {{base_url}}/api/accounts \
  -H "Authorization: Bearer {{jwt}}"
```

### Create Account
```bash
curl -X POST {{base_url}}/api/accounts \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "HDFC Bank",
    "initialBalance": 50000.00
  }'
```

### Update Account
```bash
curl -X PUT {{base_url}}/api/accounts/ACCOUNT_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "HDFC Savings Account",
    "initialBalance": 55000.00
  }'
```

### Delete Account
```bash
curl -X DELETE {{base_url}}/api/accounts/ACCOUNT_ID \
  -H "Authorization: Bearer {{jwt}}"
```

### Get Account Summary
```bash
curl {{base_url}}/api/accounts/summary \
  -H "Authorization: Bearer {{jwt}}"
```

### Get Account Expenses
```bash
# Basic
curl {{base_url}}/api/accounts/ACCOUNT_ID/expenses \
  -H "Authorization: Bearer {{jwt}}"

# With filters (year, month, pagination)
curl "{{base_url}}/api/accounts/ACCOUNT_ID/expenses?year=2025&month=11&page=1&limit=20" \
  -H "Authorization: Bearer {{jwt}}"
```

---

## Expenses

### Get All Expenses
```bash
# Basic
curl {{base_url}}/api/expenses \
  -H "Authorization: Bearer {{jwt}}"

# With filters
curl "{{base_url}}/api/expenses?year=2025&month=11&accountId=ACCOUNT_ID&page=1&limit=20" \
  -H "Authorization: Bearer {{jwt}}"
```

### Create Expense
```bash
# Without tags (defaults to ["misc"])
curl -X POST {{base_url}}/api/expenses \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1500.00,
    "categoryId": "CATEGORY_ID",
    "accountId": "ACCOUNT_ID",
    "date": "2025-11-02T10:30:00Z",
    "description": "Lunch at restaurant"
  }'

# With single tag (including merchant name in tag)
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

# With multiple tags (merchant, category, etc)
curl -X POST {{base_url}}/api/expenses \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1500.00,
    "categoryId": "CATEGORY_ID",
    "accountId": "ACCOUNT_ID",
    "date": "2025-11-02T10:30:00Z",
    "description": "Business lunch with client at Starbucks",
    "tags": ["food", "Starbucks", "business", "client-meeting", "reimbursable"]
  }'
```

### Update Expense
```bash
# Update amount and tags
curl -X PUT {{base_url}}/api/expenses/EXPENSE_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1650.00,
    "tags": ["food", "Subway", "dining", "tip-included"]
  }'

# Update only tags (e.g., changing merchant)
curl -X PUT {{base_url}}/api/expenses/EXPENSE_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["food", "Chipotle", "lunch"]
  }'
```

### Delete Expense
```bash
curl -X DELETE {{base_url}}/api/expenses/EXPENSE_ID \
  -H "Authorization: Bearer {{jwt}}"
```

### Get All Expense Tags
Get all unique tags from user's expenses, sorted alphabetically.

```bash
curl {{base_url}}/api/expenses/tags \
  -H "Authorization: Bearer {{jwt}}"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "tags": ["business", "food", "misc", "shopping", "transport"],
    "count": 5
  }
}
```

---

## Merchant Patterns

### Get All Merchant Patterns
```bash
curl {{base_url}}/api/merchant-patterns \
  -H "Authorization: Bearer {{jwt}}"
```

### Create Merchant Pattern
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

**Match Types:**
- `exact` - Exact match
- `contains` - Contains the merchant name
- `startswith` - Starts with the merchant name
- `regex` - Regular expression match

### Update Merchant Pattern
```bash
curl -X PUT {{base_url}}/api/merchant-patterns/PATTERN_ID \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "merchantName": "Swiggy*",
    "categoryId": "CATEGORY_ID",
    "matchType": "contains",
    "isActive": true
  }'
```

### Delete Merchant Pattern
```bash
curl -X DELETE {{base_url}}/api/merchant-patterns/PATTERN_ID \
  -H "Authorization: Bearer {{jwt}}"
```

### Match Merchant Pattern
```bash
curl -X POST {{base_url}}/api/merchant-patterns/match \
  -H "Authorization: Bearer {{jwt}}" \
  -H "Content-Type: application/json" \
  -d '{
    "merchantName": "SWIGGY BANGALORE"
  }'
```

---

## Complete Workflow Example

```bash
# 1. Register a user
REGISTER_RESPONSE=$(curl -s -X POST {{base_url}}/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "demo123456",
    "name": "Demo User"
  }')

# Extract token
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.data.token')

echo "Token: $TOKEN"

# 2. Create a category
CATEGORY_RESPONSE=$(curl -s -X POST {{base_url}}/api/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Food",
    "color": "#FF5722"
  }')

CATEGORY_ID=$(echo $CATEGORY_RESPONSE | jq -r '.data.id')
echo "Category ID: $CATEGORY_ID"

# 3. Create an account
ACCOUNT_RESPONSE=$(curl -s -X POST {{base_url}}/api/accounts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Main Account",
    "initialBalance": 100000.00
  }')

ACCOUNT_ID=$(echo $ACCOUNT_RESPONSE | jq -r '.data.id')
echo "Account ID: $ACCOUNT_ID"

# 4. Create an expense
curl -s -X POST {{base_url}}/api/expenses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"amount\": 350.00,
    \"categoryId\": \"$CATEGORY_ID\",
    \"accountId\": \"$ACCOUNT_ID\",
    \"date\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"description\": \"Coffee and snacks\",
    \"merchantName\": \"Starbucks\"
  }" | jq .

# 5. Get all expenses
curl -s {{base_url}}/api/expenses \
  -H "Authorization: Bearer $TOKEN" | jq .

# 6. Get account summary
curl -s {{base_url}}/api/accounts/summary \
  -H "Authorization: Bearer $TOKEN" | jq .
```

---

## Notes

1. **Authentication**: All endpoints except `/health`, `/api/auth/register`, `/api/auth/login`, and `/api/auth/refresh` require a JWT token in the Authorization header.

2. **Token Format**: `Authorization: Bearer {{jwt}}`

3. **Date Format**: ISO 8601 format (e.g., `2025-11-02T10:30:00Z`)

4. **UUIDs**: Replace placeholders like `CATEGORY_ID`, `ACCOUNT_ID`, `EXPENSE_ID`, etc. with actual UUIDs from your responses.

5. **Pagination**: Use `page` and `limit` query parameters (default: page=1, limit=50)

6. **Filtering**:
   - Expenses can be filtered by `year`, `month`, and `accountId`
   - Account expenses can be filtered by `year` and `month`

7. **Tags**:
   - Each expense can have multiple tags (array of strings)
   - If no tags are provided when creating an expense, it defaults to `["misc"]`
   - Tags are returned in sorted alphabetical order from `/api/expenses/tags`
   - Use tags for flexible categorization including:
     - Merchant names (e.g., "Subway", "Starbucks", "Amazon")
     - Expense types (e.g., "food", "transport", "shopping")
     - Additional attributes (e.g., "urgent", "reimbursable", "tax-deductible", "business")
   - Tags replace the previous `merchantName` field - include merchant names as tags

8. **Pretty Print JSON**: Add `| jq .` to any curl command to format the JSON response (requires `jq` to be installed)
