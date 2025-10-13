# Backend API Contract

This document defines the API endpoints and data structures for the Expense Tracker backend built in **Golang**.

## Overview

The backend is a **RESTful API service** built with Golang and PostgreSQL, providing real-time data synchronization and storage for the Android expense tracker app. The Android app handles:
- Reading bank notifications/SMS messages
- Parsing transaction data (amount, date, merchant, account)
- Pattern-based categorization using user-defined merchant mappings
- Local SQLite storage for offline-first operation
- Real-time HTTP sync to backend when online

The backend provides:
- User authentication and account management (JWT-based)
- RESTful API endpoints for all CRUD operations
- Real-time and batch sync capabilities
- Conflict resolution for offline changes
- Data persistence in PostgreSQL

## Data Structures

### `Category`

Represents a spending category (user-defined or system default).

| Field   | Type   | Description                      | Example        |
|---------|--------|----------------------------------|----------------|
| `id`    | string | Unique identifier for the category | "cat-1"        |
| `userId` | string | User ID (null for system defaults) | "user-123"    |
| `name`  | string | Display name of the category     | "Groceries"    |
| `color` | string | Hex color code for UI elements   | "#FFD700"      |
| `isDefault` | boolean | Whether this is a system default category | false |

### `Account`

Represents a user's bank account with balance tracking.

| Field         | Type   | Description                           | Example        |
|---------------|--------|---------------------------------------|----------------|
| `id`          | string | Unique identifier for the account     | "acc-1"        |
| `userId`      | string | User ID who owns this account         | "user-123"     |
| `name`        | string | Display name of the account          | "Main Bank"    |
| `initialBalance` | number | The initial balance of the account    | 1000           |
| `currentBalance` | number | Current balance after expenses        | 750            |
| `totalSpent`  | number | Total amount spent from this account  | 250            |

### `Expense`

Represents a single expense entry (can be manually entered or auto-detected from notifications).

| Field         | Type   | Description                             | Example                  |
|---------------|--------|-----------------------------------------|--------------------------|
| `id`          | string | Unique identifier for the expense       | "exp-123"                |
| `userId`      | string | User ID who owns this expense          | "user-123"               |
| `amount`      | number | The monetary value of the expense       | 15.75                    |
| `categoryId`  | string | ID of the associated category           | "cat-1"                  |
| `accountId`   | string | ID of the associated account           | "acc-1"                  |
| `date`        | string | ISO 8601 timestamp of the transaction   | "2025-09-16T10:00:00.000Z" |
| `description` | string | Optional note about the expense         | "Weekly groceries"       |
| `source`      | string | Source of entry: "manual" or "auto"     | "auto"                   |
| `merchantId`  | string | ID of the merchant (if detected)        | "merch-456"              |
| `merchantName` | string | Name of merchant (parsed from notification) | "Amazon"          |
| `locationId`  | string | ID of location (fallback if no merchant)| "loc-789"               |
| `rawData`     | string | Original notification text (if auto)    | "Spent Rs.15.75 at Store"|
| `verified`    | boolean| User has verified/corrected the entry   | true                     |
| `createdAt`   | string | When expense was created                | "2025-09-16T10:00:00.000Z" |
| `updatedAt`   | string | Last modification timestamp             | "2025-09-16T10:05:00.000Z" |

### `Transaction`

Represents raw transaction data captured from notifications/SMS before being converted to an expense.

| Field          | Type   | Description                              | Example                  |
|----------------|--------|------------------------------------------|--------------------------|
| `id`           | string | Unique identifier                        | "txn-001"                |
| `userId`       | string | User ID who owns this transaction       | "user-123"               |
| `rawText`      | string | Full notification/SMS text               | "Your A/C XX1234 debited by Rs.150.00..." |
| `timestamp`    | string | When notification was received           | "2025-09-16T10:00:00.000Z" |
| `senderInfo`   | string | Notification sender/package name         | "com.bank.app" or "BK-HDFC"|
| `amount`       | number | Extracted amount                         | 150.00                   |
| `merchantName` | string | Extracted merchant name (if found)       | "Amazon"                 |
| `accountLast4` | string | Last 4 digits of account number          | "1234"                   |
| `parsed`       | boolean| Whether successfully parsed              | true                     |
| `processed`    | boolean| Whether converted to expense             | true                     |
| `expenseId`    | string | Linked expense ID (after processing)     | "exp-123"                |
| `createdAt`    | string | When transaction was created             | "2025-09-16T10:00:00.000Z" |

### `MerchantInfo`

Represents merchant information detected from transactions.

| Field         | Type   | Description                              | Example                  |
|---------------|--------|------------------------------------------|--------------------------|
| `id`          | string | Unique identifier for the merchant       | "merch-456"              |
| `userId`      | string | User ID who owns this merchant record   | "user-123"               |
| `name`        | string | Merchant name                            | "Walmart Supercenter"    |
| `aliases`     | array  | Known name variations                    | ["WALMART", "WAL-MART"]  |
| `commonCategoryId` | string | Most common category for this merchant | "cat-1"            |
| `transactionCount` | number | Number of transactions with this merchant | 42                 |
| `totalSpent`  | number | Total amount spent at this merchant      | 1250.00                  |

### `MerchantPattern`

Represents user-defined rules for automatic categorization of merchants. This is the "learning system" where users teach the app how to categorize transactions.

| Field         | Type   | Description                              | Example                  |
|---------------|--------|------------------------------------------|--------------------------|
| `id`          | string | Unique identifier for the pattern        | "pat-001"                |
| `userId`      | string | User ID who owns this pattern           | "user-123"               |
| `merchantName` | string | Merchant name to match (case-insensitive) | "Amazon"                |
| `categoryId`  | string | Category to auto-assign                  | "cat-3"                  |
| `matchType`   | string | "exact" or "contains"                    | "contains"               |
| `isActive`    | boolean| Whether this pattern is currently active | true                     |
| `createdAt`   | string | When pattern was created                 | "2025-09-16T10:00:00.000Z" |
| `lastUsedAt`  | string | Last time pattern was applied            | "2025-09-20T15:30:00.000Z" |
| `useCount`    | number | Number of times pattern has been applied | 15                       |

### `Location`

Represents location data recorded when merchant detection fails (fallback mechanism).

| Field         | Type   | Description                              | Example                  |
|---------------|--------|------------------------------------------|--------------------------|
| `id`          | string | Unique identifier for the location       | "loc-789"                |
| `userId`      | string | User ID who owns this location          | "user-123"               |
| `latitude`    | number | GPS latitude                             | 37.7749                  |
| `longitude`   | number | GPS longitude                            | -122.4194                |
| `timestamp`   | string | When location was captured               | "2025-09-16T10:00:00.000Z" |
| `address`     | string | Reverse-geocoded address (if available)  | "123 Main St, SF, CA"    |
| `accuracy`    | number | Location accuracy in meters              | 15.5                     |

### `User`

Represents a user account.

| Field         | Type   | Description                              | Example                  |
|---------------|--------|------------------------------------------|--------------------------|
| `id`          | string | Unique identifier for the user           | "user-123"               |
| `email`       | string | User's email address                     | "user@example.com"       |
| `name`        | string | User's display name                      | "John Doe"               |
| `createdAt`   | string | Account creation timestamp               | "2025-09-16T10:00:00.000Z" |
| `lastLoginAt` | string | Last login timestamp                     | "2025-09-20T15:30:00.000Z" |

### `SyncStatus`

Tracks real-time synchronization status between Android app and backend.

| Field           | Type   | Description                              | Example                  |
|-----------------|--------|------------------------------------------|--------------------------|
| `id`            | string | Unique identifier                        | "sync-001"               |
| `userId`        | string | User ID                                  | "user-123"               |
| `deviceId`      | string | Device identifier                        | "pixel9-abc123"          |
| `deviceName`    | string | Human-readable device name               | "My Pixel 9"             |
| `lastSyncTime`  | string | Last successful sync timestamp           | "2025-09-16T23:00:00.000Z" |
| `lastSyncType`  | string | "realtime", "batch", or "manual"         | "realtime"               |
| `pendingCount`  | number | Number of items pending sync             | 15                       |
| `syncedCount`   | number | Number of items synced in last operation | 42                       |
| `status`        | string | "idle", "syncing", "success", "error"    | "success"                |
| `errorMessage`  | string | Error details if sync failed             | null                     |
| `conflictsResolved` | number | Number of conflicts resolved in last sync | 2                  |

---

## API Endpoints

### Categories

#### `GET /api/categories`

Retrieves a list of all available expense categories (system defaults + user custom categories).

- **Response `200 OK`**
  ```json
  [
    { "id": "cat-1", "userId": null, "name": "Groceries", "color": "#4CAF50", "isDefault": true },
    { "id": "cat-2", "userId": null, "name": "Transport", "color": "#2196F3", "isDefault": true },
    { "id": "cat-3", "userId": "user-123", "name": "Gaming", "color": "#FFC107", "isDefault": false }
  ]
  ```

#### `POST /api/categories`

Creates a new custom category for the user.

- **Request Body:**
  ```json
  {
    "name": "Pet Care",
    "color": "#8BC34A"
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "id": "cat-5",
    "userId": "user-123",
    "name": "Pet Care",
    "color": "#8BC34A",
    "isDefault": false
  }
  ```

#### `PUT /api/categories/:id`

Updates an existing custom category (user can only update their own categories, not system defaults).

- **Request Body:**
  ```json
  {
    "name": "Pet Supplies",
    "color": "#4CAF50"
  }
  ```

- **Response `200 OK`**
  - Returns the updated category object

- **Response `403 Forbidden`**
  - If trying to update a system default category

#### `DELETE /api/categories/:id`

Deletes a custom category (cannot delete system defaults).

- **Response `204 No Content`**
  - Successfully deleted

- **Response `403 Forbidden`**
  - If trying to delete a system default category or category with existing expenses

### Accounts

#### `GET /api/accounts`

Retrieves a list of all user accounts with balance information.

- **Response `200 OK`**
  ```json
  [
    { 
      "id": "acc-1", 
      "name": "Main Bank", 
      "initialBalance": 1000,
      "currentBalance": 750,
      "totalSpent": 250
    },
    { 
      "id": "acc-2", 
      "name": "Cash", 
      "initialBalance": 200,
      "currentBalance": 150,
      "totalSpent": 50
    }
  ]
  ```

#### `POST /api/accounts`

Creates a new account for the user.

- **Request Body:**
  ```json
  {
    "name": "Savings Account",
    "initialBalance": 5000
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "id": "acc-3",
    "userId": "user-123",
    "name": "Savings Account",
    "initialBalance": 5000,
    "currentBalance": 5000,
    "totalSpent": 0
  }
  ```

#### `PUT /api/accounts/:id`

Updates an existing account.

- **Request Body:**
  ```json
  {
    "name": "Updated Account Name",
    "initialBalance": 5500
  }
  ```

- **Response `200 OK`**
  - Returns the updated account object

#### `DELETE /api/accounts/:id`

Deletes an account (only if no expenses associated with it).

- **Response `204 No Content`**
  - Successfully deleted

- **Response `400 Bad Request`**
  - If account has associated expenses

#### `GET /api/accounts/summary`

Retrieves a summary of all accounts combined.

- **Response `200 OK`**
  ```json
  {
    "totalInitialBalance": 1200,
    "totalCurrentBalance": 900,
    "totalSpent": 300,
    "accountCount": 2
  }
  ```

#### `GET /api/accounts/:id/expenses`

Retrieves expenses for a specific account with pagination.

- **Query Parameters:**
  - `month` (number, optional): The month (1-12) to filter expenses by.
  - `year` (number, optional): The year to filter expenses by.
  - `page` (number, optional): The page number for pagination.
  - `limit` (number, optional): The number of items per page.

- **Response `200 OK`**
  ```json
  {
    "expenses": [
      {
        "id": "exp-1",
        "amount": 55.20,
        "categoryId": "cat-1",
        "accountId": "acc-1",
        "date": "2025-09-15T14:30:00.000Z",
        "description": "Supermarket run"
      }
    ],
    "totalPages": 1,
    "currentPage": 1,
    "totalSpentFromAccount": 175.20
  }
  ```

---

### Expenses

#### `GET /api/expenses`

Retrieves a list of expenses with smart pagination (month-wise and year-wise). The frontend implements on-demand loading.

- **Query Parameters:**
  - `month` (number, optional): The month (1-12) to filter expenses by.
  - `year` (number, optional): The year to filter expenses by.
  - `accountId` (string, optional): Filter expenses by specific account.
  - `page` (number, optional): The page number for pagination (default: 1).
  - `limit` (number, optional): The number of items per page (default: 20).
  - If no parameters are provided, returns current month's expenses.

- **Response `200 OK`**
  ```json
  {
    "expenses": [
      {
        "id": "exp-1",
        "amount": 55.20,
        "categoryId": "cat-1",
        "accountId": "acc-1",
        "date": "2025-09-15T14:30:00.000Z",
        "description": "Supermarket run"
      },
      {
        "id": "exp-2",
        "amount": 22.00,
        "categoryId": "cat-2",
        "accountId": "acc-2",
        "date": "2025-09-15T08:00:00.000Z",
        "description": "Bus fare"
      },
      {
        "id": "exp-3",
        "amount": 120.00,
        "categoryId": "cat-4",
        "accountId": "acc-1",
        "date": "2025-09-14T18:00:00.000Z",
        "description": "Electricity bill"
      }
    ],
    "totalPages": 1,
    "currentPage": 1,
    "hasMore": false,
    "totalExpenses": 3,
    "periodSummary": {
      "month": 9,
      "year": 2025,
      "totalSpent": 197.20,
      "expenseCount": 3
    }
  }
  ```

#### `POST /api/expenses`

Creates a new expense.

- **Request Body:**
  ```json
  {
    "amount": 45.00,
    "categoryId": "cat-3",
    "accountId": "acc-1",
    "date": "2025-09-16T20:00:00.000Z",
    "description": "Movie tickets"
  }
  ```

- **Response `201 Created`**
  - Returns the newly created expense object, including its server-generated `id`.
  ```json
  {
    "id": "exp-4",
    "amount": 45.00,
    "categoryId": "cat-3",
    "accountId": "acc-1",
    "date": "2025-09-16T20:00:00.000Z",
    "description": "Movie tickets"
  }
  ```

#### `POST /api/expenses/batch`

Batch create/update expenses (used for offline sync).

- **Request Body:**
  ```json
  {
    "deviceId": "pixel9-abc123",
    "expenses": [
      {
        "id": "exp-local-001",
        "amount": 45.00,
        "categoryId": "cat-3",
        "accountId": "acc-1",
        "date": "2025-09-16T20:00:00.000Z",
        "description": "Movie tickets",
        "createdAt": "2025-09-16T20:00:05.000Z",
        "updatedAt": "2025-09-16T20:00:05.000Z"
      }
    ]
  }
  ```

- **Response `200 OK`**
  ```json
  {
    "success": true,
    "synced": 1,
    "failed": 0,
    "conflicts": 0,
    "idMappings": {
      "exp-local-001": "exp-server-456"
    }
  }
  ```

#### `PUT /api/expenses/:id`

Updates an existing expense.

- **Request Body:**
  ```json
  {
    "amount": 50.00,
    "categoryId": "cat-2",
    "description": "Updated description",
    "verified": true
  }
  ```

- **Response `200 OK`**
  - Returns the updated expense object

#### `PUT /api/expenses/:id/verify`

Marks an auto-detected expense as verified by the user (or corrects it).

- **Request Body:**
  ```json
  {
    "categoryId": "cat-2",
    "description": "Corrected description",
    "verified": true
  }
  ```

- **Response `200 OK`**
  - Returns the updated expense object.

#### `DELETE /api/expenses/:id`

Deletes an expense.

- **Response `204 No Content`**
  - Successfully deleted

---

### Transactions

#### `POST /api/transactions`

Creates a new raw transaction from notification/SMS data.

- **Request Body:**
  ```json
  {
    "rawText": "Your A/C XX1234 debited by Rs.150.00 on 16-Sep-25 at AMAZON",
    "timestamp": "2025-09-16T10:00:00.000Z",
    "senderInfo": "BK-HDFC",
    "amount": 150.00,
    "merchantName": "AMAZON"
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "id": "txn-001",
    "rawText": "Your A/C XX1234 debited by Rs.150.00 on 16-Sep-25 at AMAZON",
    "timestamp": "2025-09-16T10:00:00.000Z",
    "senderInfo": "BK-HDFC",
    "amount": 150.00,
    "merchantName": "AMAZON",
    "parsed": true,
    "processed": false,
    "expenseId": null
  }
  ```

#### `POST /api/transactions/batch`

Batch upload transactions (used during sync from local storage).

- **Request Body:**
  ```json
  {
    "deviceId": "pixel9-abc123",
    "transactions": [
      {
        "rawText": "Txn 1 text...",
        "timestamp": "2025-09-16T10:00:00.000Z",
        "senderInfo": "BK-HDFC",
        "amount": 150.00
      },
      {
        "rawText": "Txn 2 text...",
        "timestamp": "2025-09-16T11:00:00.000Z",
        "senderInfo": "BK-SBI",
        "amount": 75.50
      }
    ]
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "success": true,
    "processed": 2,
    "failed": 0,
    "transactionIds": ["txn-001", "txn-002"]
  }
  ```

#### `GET /api/transactions`

Retrieves raw transactions with filtering.

- **Query Parameters:**
  - `processed` (boolean, optional): Filter by processing status (whether converted to expense)
  - `startDate` (string, optional): Start date filter
  - `endDate` (string, optional): End date filter
  - `limit` (number, optional): Number of items per page

- **Response `200 OK`**
  ```json
  {
    "transactions": [
      {
        "id": "txn-001",
        "rawText": "Your A/C XX1234 debited by Rs.150.00 at AMAZON",
        "timestamp": "2025-09-16T10:00:00.000Z",
        "amount": 150.00,
        "merchantName": "Amazon",
        "parsed": true,
        "processed": true,
        "expenseId": "exp-123"
      }
    ],
    "total": 1
  }
  ```

---

### Merchants

#### `GET /api/merchants`

Retrieves list of detected merchants.

- **Query Parameters:**
  - `search` (string, optional): Search by merchant name
  - `categoryId` (string, optional): Filter by category

- **Response `200 OK`**
  ```json
  [
    {
      "id": "merch-456",
      "name": "Walmart Supercenter",
      "aliases": ["WALMART", "WAL-MART"],
      "commonCategoryId": "cat-1",
      "transactionCount": 42,
      "totalSpent": 1250.00
    }
  ]
  ```

#### `POST /api/merchants`

Creates or updates merchant information.

- **Request Body:**
  ```json
  {
    "name": "Amazon India",
    "aliases": ["AMAZON", "AMZ"]
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "id": "merch-789",
    "name": "Amazon India",
    "aliases": ["AMAZON", "AMZ"],
    "commonCategoryId": null,
    "transactionCount": 0,
    "totalSpent": 0
  }
  ```

#### `GET /api/merchants/:id/expenses`

Get all expenses associated with a specific merchant.

- **Response `200 OK`**
  ```json
  {
    "merchant": {
      "id": "merch-456",
      "name": "Amazon India"
    },
    "expenses": [...],
    "totalSpent": 1250.00,
    "expenseCount": 15
  }
  ```

---

### Merchant Patterns

Merchant Patterns are the core of the user-learning system. When a user categorizes a merchant, they can create a pattern so future transactions from that merchant are automatically categorized.

#### `GET /api/merchant-patterns`

Retrieves all merchant patterns for the authenticated user.

- **Query Parameters:**
  - `isActive` (boolean, optional): Filter by active/inactive patterns

- **Response `200 OK`**
  ```json
  [
    {
      "id": "pat-001",
      "merchantName": "Amazon",
      "categoryId": "cat-3",
      "matchType": "contains",
      "isActive": true,
      "createdAt": "2025-09-16T10:00:00.000Z",
      "lastUsedAt": "2025-09-20T15:30:00.000Z",
      "useCount": 15
    },
    {
      "id": "pat-002",
      "merchantName": "Uber",
      "categoryId": "cat-2",
      "matchType": "exact",
      "isActive": true,
      "createdAt": "2025-09-17T12:00:00.000Z",
      "lastUsedAt": "2025-09-19T08:30:00.000Z",
      "useCount": 8
    }
  ]
  ```

#### `POST /api/merchant-patterns`

Creates a new merchant pattern. This is typically called when a user chooses "Always categorize [Merchant] as [Category]".

- **Request Body:**
  ```json
  {
    "merchantName": "Starbucks",
    "categoryId": "cat-5",
    "matchType": "contains"
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "id": "pat-003",
    "userId": "user-123",
    "merchantName": "Starbucks",
    "categoryId": "cat-5",
    "matchType": "contains",
    "isActive": true,
    "createdAt": "2025-09-21T10:00:00.000Z",
    "lastUsedAt": null,
    "useCount": 0
  }
  ```

- **Response `409 Conflict`**
  - If a pattern for this merchant already exists

#### `PUT /api/merchant-patterns/:id`

Updates an existing merchant pattern (e.g., change category or match type).

- **Request Body:**
  ```json
  {
    "categoryId": "cat-4",
    "matchType": "exact",
    "isActive": true
  }
  ```

- **Response `200 OK`**
  - Returns the updated pattern object

#### `DELETE /api/merchant-patterns/:id`

Deletes a merchant pattern.

- **Response `204 No Content`**
  - Successfully deleted

#### `POST /api/merchant-patterns/match`

Tests which pattern (if any) would match a given merchant name. Used by the Android app for client-side categorization.

- **Request Body:**
  ```json
  {
    "merchantName": "AMAZON INDIA"
  }
  ```

- **Response `200 OK`**
  ```json
  {
    "matched": true,
    "pattern": {
      "id": "pat-001",
      "merchantName": "Amazon",
      "categoryId": "cat-3",
      "matchType": "contains"
    }
  }
  ```

- **Response `200 OK` (no match)**
  ```json
  {
    "matched": false,
    "pattern": null
  }
  ```

---

### Locations

#### `POST /api/locations`

Records location data (fallback when merchant not detected).

- **Request Body:**
  ```json
  {
    "latitude": 37.7749,
    "longitude": -122.4194,
    "timestamp": "2025-09-16T10:00:00.000Z",
    "accuracy": 15.5
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "id": "loc-789",
    "latitude": 37.7749,
    "longitude": -122.4194,
    "timestamp": "2025-09-16T10:00:00.000Z",
    "address": "123 Main St, SF, CA",
    "accuracy": 15.5
  }
  ```

#### `GET /api/locations/:id/expenses`

Get expenses associated with a location.

- **Response `200 OK`**
  ```json
  {
    "location": {
      "id": "loc-789",
      "address": "123 Main St, SF, CA"
    },
    "expenses": [...],
    "totalSpent": 350.00
  }
  ```

---

### Sync Operations

The sync system supports both real-time and batch synchronization. Real-time sync occurs when the device is online, while batch sync handles offline changes when the device reconnects.

#### `POST /api/sync/incremental`

Performs an incremental sync of all changes since last sync. Supports transactions, expenses, categories, and merchant patterns.

- **Request Body:**
  ```json
  {
    "deviceId": "pixel9-abc123",
    "lastSyncTimestamp": "2025-09-20T10:00:00.000Z",
    "changes": {
      "transactions": [
        {
          "id": "txn-local-001",
          "rawText": "Debited Rs.150.00 at Amazon",
          "timestamp": "2025-09-20T11:00:00.000Z",
          "amount": 150.00,
          "merchantName": "Amazon",
          "createdAt": "2025-09-20T11:00:05.000Z"
        }
      ],
      "expenses": [
        {
          "id": "exp-local-001",
          "amount": 150.00,
          "categoryId": "cat-3",
          "accountId": "acc-1",
          "date": "2025-09-20T11:00:00.000Z",
          "merchantName": "Amazon",
          "verified": true,
          "createdAt": "2025-09-20T11:01:00.000Z",
          "updatedAt": "2025-09-20T11:01:00.000Z"
        }
      ],
      "merchantPatterns": [
        {
          "id": "pat-local-001",
          "merchantName": "Amazon",
          "categoryId": "cat-3",
          "matchType": "contains",
          "createdAt": "2025-09-20T11:02:00.000Z"
        }
      ]
    }
  }
  ```

- **Response `200 OK`**
  ```json
  {
    "success": true,
    "syncTimestamp": "2025-09-20T12:00:00.000Z",
    "conflicts": [],
    "idMappings": {
      "transactions": {
        "txn-local-001": "txn-server-456"
      },
      "expenses": {
        "exp-local-001": "exp-server-789"
      },
      "merchantPatterns": {
        "pat-local-001": "pat-server-123"
      }
    },
    "serverChanges": {
      "transactions": [],
      "expenses": [],
      "categories": [],
      "merchantPatterns": []
    }
  }
  ```

- **Response `200 OK` (with conflicts)**
  ```json
  {
    "success": true,
    "syncTimestamp": "2025-09-20T12:00:00.000Z",
    "conflicts": [
      {
        "type": "expense",
        "localId": "exp-local-002",
        "serverId": "exp-server-002",
        "resolution": "server_wins",
        "reason": "Server version is newer"
      }
    ],
    "idMappings": {...},
    "serverChanges": {...}
  }
  ```

#### `POST /api/sync/batch/expenses`

Batch sync expenses (used for offline sync when device reconnects).

- **Request Body:**
  ```json
  {
    "deviceId": "pixel9-abc123",
    "expenses": [
      {
        "id": "exp-local-001",
        "amount": 45.00,
        "categoryId": "cat-3",
        "accountId": "acc-1",
        "date": "2025-09-19T14:00:00.000Z",
        "description": "Offline expense",
        "createdAt": "2025-09-19T14:00:05.000Z",
        "updatedAt": "2025-09-19T14:00:05.000Z"
      }
    ]
  }
  ```

- **Response `200 OK`**
  ```json
  {
    "success": true,
    "synced": 1,
    "failed": 0,
    "conflicts": 0,
    "idMappings": {
      "exp-local-001": "exp-server-890"
    }
  }
  ```

#### `GET /api/sync/status/:deviceId`

Gets current sync status for a device.

- **Response `200 OK`**
  ```json
  {
    "id": "sync-001",
    "userId": "user-123",
    "deviceId": "pixel9-abc123",
    "deviceName": "My Pixel 9",
    "lastSyncTime": "2025-09-20T12:00:00.000Z",
    "lastSyncType": "realtime",
    "pendingCount": 0,
    "syncedCount": 42,
    "status": "success",
    "errorMessage": null,
    "conflictsResolved": 0
  }
  ```

#### `POST /api/sync/status`

Updates sync status from device (called after successful sync).

- **Request Body:**
  ```json
  {
    "deviceId": "pixel9-abc123",
    "syncType": "batch",
    "syncedCount": 15,
    "pendingCount": 0,
    "conflictsResolved": 2
  }
  ```

- **Response `200 OK`**
  ```json
  {
    "id": "sync-001",
    "deviceId": "pixel9-abc123",
    "lastSyncTime": "2025-09-20T12:00:00.000Z",
    "status": "success"
  }
  ```

---

## Data Flow

### Automated Transaction Processing

1. **Notification Capture** (Android App - Local)
   - Notification listener service detects bank SMS/notification
   - Parse transaction details: amount, merchant name, account last 4 digits, timestamp
   - Create `Transaction` record in local SQLite database (encrypted)
   - Store raw notification text for user reference

2. **Pattern-Based Categorization** (Android App - Local)
   - Check if merchant name matches any existing `MerchantPattern`
   - If match found:
     - Auto-create `Expense` with the pattern's category
     - Link expense to transaction
     - Mark as verified=false (user can review later)
   - If no match found:
     - Flag transaction for manual categorization
     - Optionally capture GPS location as fallback
     - Notify user to categorize

3. **Real-Time Sync to Backend** (Android App → Backend API)
   - When online: Immediately sync new transactions/expenses via HTTP API
   - Call `POST /api/sync/incremental` with changes
   - Backend returns server-generated IDs (replace local temp IDs)
   - Update local database with server IDs

4. **Offline Handling** (Android App - Local)
   - When offline: Queue all changes in local database
   - Mark records as "pending sync"
   - Continue normal operation with local data

5. **Offline Sync on Reconnect** (Android App → Backend API)
   - Detect internet connection restored
   - Call `POST /api/sync/incremental` with all pending changes since last sync
   - Backend resolves conflicts (last-write-wins based on `updatedAt` timestamp)
   - Update local database with conflict resolutions and new server IDs

6. **User Learning System** (Android App - Local & Backend)
   - User categorizes an uncategorized transaction
   - App prompts: "Always categorize [Merchant] as [Category]?"
   - If yes: Create `MerchantPattern` via `POST /api/merchant-patterns`
   - Pattern synced to backend for cross-device availability
   - Future transactions auto-categorized using this pattern

### Manual Entry Flow

1. **Add Expense** (Android App)
   - User manually enters expense details
   - Store in local SQLite with source="manual"
   - If online: Immediately sync via `POST /api/expenses`
   - If offline: Mark as pending sync

2. **Sync to Backend** (Android App → Backend)
   - Real-time: Syncs immediately when created (if online)
   - Batch: Syncs when device reconnects (if was offline)

3. **Cross-Device Access** (Android App ← Backend)
   - Other devices pull changes via `POST /api/sync/incremental`
   - Backend returns all changes since device's last sync timestamp
   - Local database updated with new expenses

### Conflict Resolution Strategy

The backend uses **last-write-wins with timestamps** for conflict resolution:

1. **Scenario**: Same expense modified on two devices while offline
2. **Resolution**: Compare `updatedAt` timestamps
3. **Winner**: Most recent timestamp wins
4. **Loser**: Overwritten, but conflict logged in sync response
5. **User Notification**: App can optionally notify user of overwritten changes

---

## Authentication

All endpoints require authentication via JWT tokens (except for auth endpoints).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

### Authentication Endpoints

#### `POST /api/auth/register`

Registers a new user account.

- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "securePassword123",
    "name": "John Doe"
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "name": "John Doe",
      "createdAt": "2025-09-16T10:00:00.000Z"
    },
    "token": "eyJhbGc..."
  }
  ```

- **Response `400 Bad Request`**
  - If email already exists or validation fails

#### `POST /api/auth/login`

Authenticates an existing user.

- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "securePassword123"
  }
  ```

- **Response `200 OK`**
  ```json
  {
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "name": "John Doe",
      "lastLoginAt": "2025-09-20T15:30:00.000Z"
    },
    "token": "eyJhbGc..."
  }
  ```

- **Response `401 Unauthorized`**
  - If email/password combination is invalid

#### `POST /api/auth/refresh`

Refreshes an expired or soon-to-expire JWT token.

- **Request Body:**
  ```json
  {
    "token": "eyJhbGc..."
  }
  ```

- **Response `200 OK`**
  ```json
  {
    "token": "eyJhbGc...",
    "expiresAt": "2025-09-21T15:30:00.000Z"
  }
  ```

#### `GET /api/auth/me`

Gets the current authenticated user's profile.

- **Headers:** Requires `Authorization: Bearer <token>`

- **Response `200 OK`**
  ```json
  {
    "id": "user-123",
    "email": "user@example.com",
    "name": "John Doe",
    "createdAt": "2025-09-16T10:00:00.000Z",
    "lastLoginAt": "2025-09-20T15:30:00.000Z"
  }
  ```

#### `POST /api/auth/devices/register`

Registers a new device for the authenticated user.

- **Headers:** Requires `Authorization: Bearer <token>`

- **Request Body:**
  ```json
  {
    "deviceId": "pixel9-abc123",
    "deviceName": "My Pixel 9"
  }
  ```

- **Response `201 Created`**
  ```json
  {
    "id": "device-001",
    "deviceId": "pixel9-abc123",
    "deviceName": "My Pixel 9",
    "userId": "user-123",
    "registeredAt": "2025-09-16T10:00:00.000Z"
  }
  ```