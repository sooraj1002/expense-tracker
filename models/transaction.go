package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	RawText      string     `json:"rawText" db:"raw_text" binding:"required"`
	Timestamp    time.Time  `json:"timestamp" db:"timestamp" binding:"required"`
	SenderInfo   string     `json:"senderInfo,omitempty" db:"sender_info"`
	Amount       *float64   `json:"amount,omitempty" db:"amount"`
	MerchantName string     `json:"merchantName,omitempty" db:"merchant_name"`
	AccountLast4 string     `json:"accountLast4,omitempty" db:"account_last4"`
	Parsed       bool       `json:"parsed" db:"parsed"`
	Processed    bool       `json:"processed" db:"processed"`
	ExpenseID    *uuid.UUID `json:"expenseId,omitempty" db:"expense_id"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

type CreateTransactionRequest struct {
	RawText      string    `json:"rawText" binding:"required"`
	Timestamp    time.Time `json:"timestamp" binding:"required"`
	SenderInfo   string    `json:"senderInfo"`
	Amount       float64   `json:"amount"`
	MerchantName string    `json:"merchantName"`
}

type BatchTransactionRequest struct {
	DeviceID     string        `json:"deviceId" binding:"required"`
	Transactions []Transaction `json:"transactions" binding:"required"`
}

type BatchTransactionResponse struct {
	Success        bool     `json:"success"`
	Processed      int      `json:"processed"`
	Failed         int      `json:"failed"`
	TransactionIDs []string `json:"transactionIds"`
}

type TransactionListResponse struct {
	Transactions []Transaction `json:"transactions"`
	Total        int           `json:"total"`
}
