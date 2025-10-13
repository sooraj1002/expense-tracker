package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Email       string     `json:"email" db:"email" binding:"required,email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name        string     `json:"name" db:"name" binding:"required"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
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
