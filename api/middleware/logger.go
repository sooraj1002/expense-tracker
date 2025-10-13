package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sooraj1002/expense-tracker/logger"
)

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(startTime)
		logger.Log.Infow("HTTP Request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration.String(),
			"ip", c.ClientIP(),
		)
	}
}
