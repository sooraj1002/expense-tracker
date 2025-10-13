package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/config"
	"github.com/sooraj1002/expense-tracker/models"
	"github.com/sooraj1002/expense-tracker/utils"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Authorization header required",
			))
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid authorization header format",
			))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(tokenString, config.AppConfig.JWT.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid or expired token",
			))
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, ErrUserIDNotFound
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrInvalidUserID
	}

	return uid, nil
}

var (
	ErrUserIDNotFound = gin.Error{Err: nil, Type: gin.ErrorTypePrivate, Meta: "user ID not found in context"}
	ErrInvalidUserID  = gin.Error{Err: nil, Type: gin.ErrorTypePrivate, Meta: "invalid user ID in context"}
)
