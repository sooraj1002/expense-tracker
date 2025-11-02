package db

import (
	"fmt"

	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
)

// RunMigrations executes GORM AutoMigrate for all models
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database not initialized, call InitDB first")
	}

	logger.Log.Info("Running database migrations...")

	// Run AutoMigrate for all models
	err := DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Account{},
		&models.Expense{},
		&models.MerchantPattern{},
		&models.MerchantInfo{},
		&models.Device{},
		&models.SyncStatus{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Log.Info("All migrations completed successfully")
	return nil
}
