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

// CategoryBudget stores a user's recurring monthly budget for a category.
// It is separate from Category because default categories are shared by every user.
type CategoryBudget struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `json:"userId" gorm:"type:uuid;column:user_id;not null;uniqueIndex:idx_category_budget_user_category;index"`
	CategoryID uuid.UUID `json:"categoryId" gorm:"type:uuid;column:category_id;not null;uniqueIndex:idx_category_budget_user_category;index"`
	Amount     float64   `json:"amount" gorm:"type:decimal(15,2);not null;check:amount > 0"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	User       User      `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Category   Category  `json:"-" gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// CategoryResponse adds the authenticated user's budget status to a category.
type CategoryResponse struct {
	Category
	MonthlyBudget *float64 `json:"monthlyBudget"`
	MonthlySpent  float64  `json:"monthlySpent"`
}

type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required,len=7"`
}

type UpdateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required,len=7"`
}

type SetCategoryBudgetRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}
