package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type MerchantInfo struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           uuid.UUID      `json:"userId" gorm:"type:uuid;column:user_id;not null;index"`
	Name             string         `json:"name" gorm:"not null" binding:"required"`
	Aliases          pq.StringArray `json:"aliases" gorm:"type:text[]"`
	CommonCategoryID *uuid.UUID     `json:"commonCategoryId,omitempty" gorm:"type:uuid;column:common_category_id"`
	TransactionCount int            `json:"transactionCount" gorm:"column:transaction_count;default:0"`
	TotalSpent       float64        `json:"totalSpent" gorm:"column:total_spent;type:decimal(15,2);default:0"`
	CreatedAt        time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
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
