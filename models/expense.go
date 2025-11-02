package models

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID  `json:"userId" gorm:"type:uuid;column:user_id;not null;index"`
	Amount       float64    `json:"amount" gorm:"type:decimal(15,2);not null" binding:"required,gt=0"`
	CategoryID   uuid.UUID  `json:"categoryId" gorm:"type:uuid;column:category_id;not null;index" binding:"required"`
	AccountID    uuid.UUID  `json:"accountId" gorm:"type:uuid;column:account_id;not null;index" binding:"required"`
	Date         time.Time  `json:"date" gorm:"not null;index" binding:"required"`
	Description  string     `json:"description,omitempty" gorm:"type:text"`
	Source       string     `json:"source" gorm:"default:'manual'"`
	MerchantID   *uuid.UUID `json:"merchantId,omitempty" gorm:"type:uuid;column:merchant_id;index"`
	MerchantName string     `json:"merchantName,omitempty" gorm:"column:merchant_name"`
	LocationID   *uuid.UUID `json:"locationId,omitempty" gorm:"type:uuid;column:location_id"`
	RawData      string     `json:"rawData,omitempty" gorm:"column:raw_data;type:text"`
	Verified     bool       `json:"verified" gorm:"default:false"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
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
