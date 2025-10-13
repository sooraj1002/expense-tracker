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
