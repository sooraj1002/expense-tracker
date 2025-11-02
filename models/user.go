package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null" binding:"required,email"`
	PasswordHash string     `json:"-" gorm:"column:password_hash;not null"`
	Name         string     `json:"name" gorm:"not null" binding:"required"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty" gorm:"column:last_login_at"`
	UpdatedAt    time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (User) BeforeCreate(tx *gorm.DB) error {
	return nil
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}
