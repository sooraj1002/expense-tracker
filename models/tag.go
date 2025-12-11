package models

import (
	"time"

	"github.com/google/uuid"
)

// Tag is a normalized tag per user.
type Tag struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"userId" gorm:"type:uuid;not null;column:user_id;uniqueIndex:idx_user_tag_name"`
	Name      string    `json:"name" gorm:"type:text;not null;uniqueIndex:idx_user_tag_name"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ExpenseTag is the join between expenses and tags.
type ExpenseTag struct {
	ExpenseID uuid.UUID `json:"expenseId" gorm:"type:uuid;primaryKey;column:expense_id"`
	TagID     uuid.UUID `json:"tagId" gorm:"type:uuid;primaryKey;column:tag_id"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	Tag       Tag       `json:"-" gorm:"constraint:OnDelete:CASCADE;foreignKey:TagID;references:ID"`
	Expense   Expense   `json:"-" gorm:"constraint:OnDelete:CASCADE;foreignKey:ExpenseID;references:ID"`
}

func (ExpenseTag) TableName() string {
	return "expense_tags"
}
