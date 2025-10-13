package models

import (
	"time"

	"github.com/google/uuid"
)

type Location struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Latitude  float64   `json:"latitude" db:"latitude" binding:"required"`
	Longitude float64   `json:"longitude" db:"longitude" binding:"required"`
	Timestamp time.Time `json:"timestamp" db:"timestamp" binding:"required"`
	Address   string    `json:"address,omitempty" db:"address"`
	Accuracy  float64   `json:"accuracy,omitempty" db:"accuracy"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type CreateLocationRequest struct {
	Latitude  float64   `json:"latitude" binding:"required"`
	Longitude float64   `json:"longitude" binding:"required"`
	Timestamp time.Time `json:"timestamp" binding:"required"`
	Accuracy  float64   `json:"accuracy"`
}

type LocationExpensesResponse struct {
	Location   Location  `json:"location"`
	Expenses   []Expense `json:"expenses"`
	TotalSpent float64   `json:"totalSpent"`
}
