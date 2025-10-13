package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type MerchantInfo struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	UserID           uuid.UUID      `json:"userId" db:"user_id"`
	Name             string         `json:"name" db:"name" binding:"required"`
	Aliases          pq.StringArray `json:"aliases" db:"aliases"`
	CommonCategoryID *uuid.UUID     `json:"commonCategoryId,omitempty" db:"common_category_id"`
	TransactionCount int            `json:"transactionCount" db:"transaction_count"`
	TotalSpent       float64        `json:"totalSpent" db:"total_spent"`
	CreatedAt        time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time      `json:"updatedAt" db:"updated_at"`
}

type CreateMerchantRequest struct {
	Name    string   `json:"name" binding:"required"`
	Aliases []string `json:"aliases"`
}

type MerchantExpensesResponse struct {
	Merchant     MerchantInfo `json:"merchant"`
	Expenses     []Expense    `json:"expenses"`
	TotalSpent   float64      `json:"totalSpent"`
	ExpenseCount int          `json:"expenseCount"`
}
