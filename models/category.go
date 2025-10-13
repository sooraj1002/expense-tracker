package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"userId,omitempty" db:"user_id"`
	Name      string     `json:"name" db:"name" binding:"required"`
	Color     string     `json:"color" db:"color" binding:"required,len=7"`
	IsDefault bool       `json:"isDefault" db:"is_default"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required,len=7"`
}

type UpdateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required,len=7"`
}
