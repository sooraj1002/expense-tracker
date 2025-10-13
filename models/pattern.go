package models

import (
	"time"

	"github.com/google/uuid"
)

type MerchantPattern struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	MerchantName string     `json:"merchantName" db:"merchant_name" binding:"required"`
	CategoryID   uuid.UUID  `json:"categoryId" db:"category_id" binding:"required"`
	MatchType    string     `json:"matchType" db:"match_type"`
	IsActive     bool       `json:"isActive" db:"is_active"`
	UseCount     int        `json:"useCount" db:"use_count"`
	LastUsedAt   *time.Time `json:"lastUsedAt,omitempty" db:"last_used_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
}

type CreatePatternRequest struct {
	MerchantName string    `json:"merchantName" binding:"required"`
	CategoryID   uuid.UUID `json:"categoryId" binding:"required"`
	MatchType    string    `json:"matchType" binding:"required,oneof=exact contains"`
}

type UpdatePatternRequest struct {
	CategoryID *uuid.UUID `json:"categoryId"`
	MatchType  *string    `json:"matchType" binding:"omitempty,oneof=exact contains"`
	IsActive   *bool      `json:"isActive"`
}

type MatchPatternRequest struct {
	MerchantName string `json:"merchantName" binding:"required"`
}

type MatchPatternResponse struct {
	Matched bool             `json:"matched"`
	Pattern *MerchantPattern `json:"pattern"`
}
