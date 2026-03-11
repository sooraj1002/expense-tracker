package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sooraj1002/expense-tracker/config"
)

var defaultAllowedOrigins = map[string]struct{}{
	"http://localhost:3000":  {},
	"http://127.0.0.1:3000":  {},
	"https://localhost:3000": {},
}

// CORSMiddleware handles CORS for browser clients.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if allowedOrigin(origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Vary", "Origin")
		}

		requestHeaders := c.GetHeader("Access-Control-Request-Headers")
		if strings.TrimSpace(requestHeaders) == "" {
			requestHeaders = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", requestHeaders)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func allowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	allowed := defaultAllowedOrigins
	if config.AppConfig != nil && len(config.AppConfig.CORS.AllowedOrigins) > 0 {
		allowed = make(map[string]struct{}, len(config.AppConfig.CORS.AllowedOrigins))
		for _, item := range config.AppConfig.CORS.AllowedOrigins {
			allowed[item] = struct{}{}
		}
	}

	_, ok := allowed[origin]
	return ok
}
