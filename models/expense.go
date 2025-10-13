package models

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	Amount       float64    `json:"amount" db:"amount" binding:"required,gt=0"`
	CategoryID   uuid.UUID  `json:"categoryId" db:"category_id" binding:"required"`
	AccountID    uuid.UUID  `json:"accountId" db:"account_id" binding:"required"`
	Date         time.Time  `json:"date" db:"date" binding:"required"`
	Description  string     `json:"description,omitempty" db:"description"`
	Source       string     `json:"source" db:"source"`
	MerchantID   *uuid.UUID `json:"merchantId,omitempty" db:"merchant_id"`
	MerchantName string     `json:"merchantName,omitempty" db:"merchant_name"`
	LocationID   *uuid.UUID `json:"locationId,omitempty" db:"location_id"`
	RawData      string     `json:"rawData,omitempty" db:"raw_data"`
	Verified     bool       `json:"verified" db:"verified"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
}

type CreateExpenseRequest struct {
	Amount       float64   `json:"amount" binding:"required,gt=0"`
	CategoryID   uuid.UUID `json:"categoryId" binding:"required"`
	AccountID    uuid.UUID `json:"accountId" binding:"required"`
	Date         time.Time `json:"date" binding:"required"`
	Description  string    `json:"description"`
	MerchantName string    `json:"merchantName"`
}

type UpdateExpenseRequest struct {
	Amount      *float64   `json:"amount" binding:"omitempty,gt=0"`
	CategoryID  *uuid.UUID `json:"categoryId"`
	AccountID   *uuid.UUID `json:"accountId"`
	Date        *time.Time `json:"date"`
	Description *string    `json:"description"`
	Verified    *bool      `json:"verified"`
}

type BatchExpenseRequest struct {
	DeviceID string    `json:"deviceId" binding:"required"`
	Expenses []Expense `json:"expenses" binding:"required"`
}

type BatchExpenseResponse struct {
	Success    bool                `json:"success"`
	Synced     int                 `json:"synced"`
	Failed     int                 `json:"failed"`
	Conflicts  int                 `json:"conflicts"`
	IDMappings map[string]string   `json:"idMappings"`
}

type ExpenseListResponse struct {
	Expenses      []Expense      `json:"expenses"`
	TotalPages    int            `json:"totalPages"`
	CurrentPage   int            `json:"currentPage"`
	HasMore       bool           `json:"hasMore"`
	TotalExpenses int            `json:"totalExpenses"`
	PeriodSummary *PeriodSummary `json:"periodSummary,omitempty"`
}

type PeriodSummary struct {
	Month        int     `json:"month"`
	Year         int     `json:"year"`
	TotalSpent   float64 `json:"totalSpent"`
	ExpenseCount int     `json:"expenseCount"`
}
