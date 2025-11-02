package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/sooraj1002/expense-tracker/api"
	"github.com/sooraj1002/expense-tracker/config"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long:  `Start the HTTP API server to handle expense tracking requests`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		if err := config.LoadConfig(); err != nil {
			logger.Log.Fatalw("Failed to load configuration", "error", err)
		}

		logger.Log.Infow("Starting Expense Tracker API Server",
			"port", config.AppConfig.Server.Port,
			"environment", config.AppConfig.Server.Environment,
		)

		// Initialize database
		dbConfig := config.AppConfig.Database
		logger.Log.Infow("Connecting to database",
			"host", dbConfig.Host,
			"port", dbConfig.Port,
			"user", dbConfig.User,
			"dbname", dbConfig.DBName,
		)
		_, err := db.InitDB(
			dbConfig.Host,
			dbConfig.Port,
			dbConfig.User,
			dbConfig.Password,
			dbConfig.DBName,
			dbConfig.SSLMode,
		)
		if err != nil {
			logger.Log.Fatalw("Failed to initialize database", "error", err)
		}
		defer db.Close()

		// Run migrations
		if err := db.RunMigrations(); err != nil {
			logger.Log.Fatalw("Failed to run migrations", "error", err)
		}

		// Setup router
		router := api.SetupRouter()

		// Create HTTP server
		srv := &http.Server{
			Addr:    ":" + config.AppConfig.Server.Port,
			Handler: router,
		}

		// Start server in a goroutine
		go func() {
			logger.Log.Infow("Server is listening", "address", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Log.Fatalw("Failed to start server", "error", err)
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Log.Info("Shutting down server...")

		// Graceful shutdown with 5 second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Log.Fatalw("Server forced to shutdown", "error", err)
		}

		logger.Log.Info("Server exited")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
