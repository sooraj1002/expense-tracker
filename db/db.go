package db

import (
	"fmt"
	"time"

	"github.com/sooraj1002/expense-tracker/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the PostgreSQL database connection using GORM
func InitDB(host string, port int, user, password, dbName, sslMode string) (*gorm.DB, error) {
	// Build the DSN for the target database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode,
	)

	logger.Log.Infof("Connecting to database: %s", dbName)

	// Configure GORM
	config := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		// If connection fails, try to create the database
		logger.Log.Warnf("Failed to connect to %s, attempting to create database...", dbName)

		// Connect to postgres database to create the target database
		postgresDSN := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
			host, port, user, password, sslMode,
		)

		postgresDB, err := gorm.Open(postgres.Open(postgresDSN), config)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres database: %w", err)
		}

		// Get the underlying SQL DB to execute raw SQL
		sqlDB, err := postgresDB.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get database instance: %w", err)
		}
		defer sqlDB.Close()

		// Create the database
		createSQL := fmt.Sprintf("CREATE DATABASE %s", dbName)
		if err := postgresDB.Exec(createSQL).Error; err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}

		logger.Log.Infof("Database '%s' created successfully", dbName)

		// Now connect to the newly created database
		db, err = gorm.Open(postgres.Open(dsn), config)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to newly created database: %w", err)
		}
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	var dbNameCheck string
	if err := db.Raw("SELECT current_database()").Scan(&dbNameCheck).Error; err != nil {
		return nil, fmt.Errorf("failed to verify database connection: %w", err)
	}

	logger.Log.Infof("Successfully connected to database: %s", dbNameCheck)

	DB = db
	return db, nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
