package models

import (
	"time"

	"github.com/google/uuid"
)

type MerchantPattern struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID  `json:"userId" gorm:"type:uuid;column:user_id;not null;index"`
	MerchantName string     `json:"merchantName" gorm:"column:merchant_name;not null" binding:"required"`
	CategoryID   uuid.UUID  `json:"categoryId" gorm:"type:uuid;column:category_id;not null" binding:"required"`
	MatchType    string     `json:"matchType" gorm:"column:match_type;default:'exact'"`
	IsActive     bool       `json:"isActive" gorm:"column:is_active;default:true"`
	UseCount     int        `json:"useCount" gorm:"column:use_count;default:0"`
	LastUsedAt   *time.Time `json:"lastUsedAt,omitempty" gorm:"column:last_used_at"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
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
