package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sooraj1002/expense-tracker/api/handlers"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/config"
)

// SetupRouter configures all routes and middleware
func SetupRouter() *gin.Engine {
	// Set Gin mode based on environment
	if config.AppConfig.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api")
	{
		// Public auth endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/refresh", handlers.RefreshToken)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Auth - user profile and device registration
			protected.GET("/auth/me", handlers.GetMe)
			protected.POST("/auth/devices/register", handlers.RegisterDevice)

			// Categories
			protected.GET("/categories", handlers.GetCategories)
			protected.POST("/categories", handlers.CreateCategory)
			protected.PUT("/categories/:id", handlers.UpdateCategory)
			protected.DELETE("/categories/:id", handlers.DeleteCategory)

			// Accounts
			protected.GET("/accounts", handlers.GetAccounts)
			protected.POST("/accounts", handlers.CreateAccount)
			protected.PUT("/accounts/:id", handlers.UpdateAccount)
			protected.DELETE("/accounts/:id", handlers.DeleteAccount)
			protected.GET("/accounts/summary", handlers.GetAccountSummary)
			protected.GET("/accounts/:id/expenses", handlers.GetAccountExpenses)

			// Expenses
			protected.GET("/expenses", handlers.GetExpenses)
			protected.POST("/expenses", handlers.CreateExpense)
			protected.PUT("/expenses/:id", handlers.UpdateExpense)
			protected.DELETE("/expenses/:id", handlers.DeleteExpense)

			// Merchant Patterns
			protected.GET("/merchant-patterns", handlers.GetMerchantPatterns)
			protected.POST("/merchant-patterns", handlers.CreateMerchantPattern)
			protected.PUT("/merchant-patterns/:id", handlers.UpdateMerchantPattern)
			protected.DELETE("/merchant-patterns/:id", handlers.DeleteMerchantPattern)
			protected.POST("/merchant-patterns/match", handlers.MatchMerchantPattern)

			// TODO: Add remaining endpoints as needed
			// - Transactions
			// - Merchants
			// - Locations
			// - Sync operations
		}
	}

	return router
}
