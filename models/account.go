package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"userId" db:"user_id"`
	Name           string    `json:"name" db:"name" binding:"required"`
	InitialBalance float64   `json:"initialBalance" db:"initial_balance"`
	CurrentBalance float64   `json:"currentBalance" db:"current_balance"`
	TotalSpent     float64   `json:"totalSpent" db:"total_spent"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type CreateAccountRequest struct {
	Name           string  `json:"name" binding:"required"`
	InitialBalance float64 `json:"initialBalance" binding:"required,min=0"`
}

type UpdateAccountRequest struct {
	Name           string  `json:"name" binding:"required"`
	InitialBalance float64 `json:"initialBalance" binding:"required,min=0"`
}

type AccountSummary struct {
	TotalInitialBalance float64 `json:"totalInitialBalance"`
	TotalCurrentBalance float64 `json:"totalCurrentBalance"`
	TotalSpent          float64 `json:"totalSpent"`
	AccountCount        int     `json:"accountCount"`
}
