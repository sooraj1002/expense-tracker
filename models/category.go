package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    *uuid.UUID `json:"userId,omitempty" gorm:"type:uuid;column:user_id;index"`
	Name      string     `json:"name" gorm:"not null" binding:"required"`
	Color     string     `json:"color" gorm:"not null" binding:"required,len=7"`
	IsDefault bool       `json:"isDefault" gorm:"column:is_default;default:false"`
	CreatedAt time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}

type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required,len=7"`
}

type UpdateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required,len=7"`
}
