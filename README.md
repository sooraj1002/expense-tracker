# Expense Tracker Backend API

A RESTful API server built with Go for tracking personal expenses. Designed to work with an Android app that reads bank notifications and automatically categorizes transactions.

## Features

- **User Authentication** - JWT-based authentication with secure password hashing
- **Expense Management** - Create, read, update, and delete expenses
- **Account Management** - Track multiple bank accounts with balance calculations
- **Category System** - System default and user-custom categories
- **Merchant Patterns** - User-defined rules for automatic expense categorization
- **Transaction Tracking** - Store raw notification data linked to expenses
- **Sync Operations** - Real-time and batch synchronization for offline support

## Tech Stack

- **Go 1.24** - Programming language
- **Gin** - HTTP web framework
- **PostgreSQL** - Database
- **JWT** - Authentication
- **bcrypt** - Password hashing

## Quick Start

1. Clone and install dependencies:
```bash
cd expense-tracker
go mod download
```

2. Set up PostgreSQL and create database:
```bash
createdb expense_tracker
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env with your settings
```

4. Run the server:
```bash
go run main.go serve
```

Server starts at `http://localhost:8080`

## Docker & Railway

### Local Docker Compose
1. Copy the environment template: `cp .env.example .env`
2. Adjust the values if needed (the compose file overrides only `DB_HOST`, `DB_PORT`, and `DB_SSLMODE`)
3. Start the stack:
   ```bash
   docker compose up --build
   ```
   The API will be available at http://localhost:8080 once the database health check passes.

### Railway Deployment
- Add a new Railway service pointing to this repository and choose the provided `Dockerfile`
- Provision a Railway PostgreSQL add-on; the platform automatically injects `DATABASE_URL`
- Define the remaining secrets (`JWT_SECRET`, `JWT_EXPIRY`, optional `LOG_LEVEL`, etc.). Railway sets `PORT` for you
- Deploy; the container runs `expense-tracker serve`, auto-migrating the database during startup

When `DATABASE_URL` is present it takes precedence over the individual `DB_*` variables, making it compatible
with managed Railway databases out of the box.

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register user
- `POST /api/auth/login` - Login
- `GET /api/auth/me` - Get profile

### Categories
- `GET /api/categories` - List categories
- `POST /api/categories` - Create category

### Accounts
- `GET /api/accounts` - List accounts
- `POST /api/accounts` - Create account

### Expenses
- `GET /api/expenses` - List expenses
- `POST /api/expenses` - Create expense

### Merchant Patterns
- `GET /api/merchant-patterns` - List patterns
- `POST /api/merchant-patterns` - Create pattern
- `POST /api/merchant-patterns/match` - Match merchant

See [BACKEND_API.md](BACKEND_API.md) for full documentation.

## Credits

- Initial CLI structure inspired by https://dev.to/aurelievache/learning-go-by-examples-part-3-create-a-cli-app-in-go-1h43
