package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID `json:"userId" gorm:"type:uuid;column:user_id;not null;index"`
	Name           string    `json:"name" gorm:"not null" binding:"required"`
	InitialBalance float64   `json:"initialBalance" gorm:"column:initial_balance;type:decimal(15,2);default:0"`
	CurrentBalance float64   `json:"currentBalance" gorm:"column:current_balance;type:decimal(15,2);default:0"`
	TotalSpent     float64   `json:"totalSpent" gorm:"column:total_spent;type:decimal(15,2);default:0"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
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
