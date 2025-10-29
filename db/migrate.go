package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/sooraj1002/expense-tracker/logger"
)

// RunMigrations executes all migration files in the migrations directory
func RunMigrations(db *sql.DB, migrationsPath string) error {
	// Create migrations table if it doesn't exist
	logger.Log.Info("Creating schema_migrations table if not exists...")
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	logger.Log.Info("schema_migrations table ready")

	// Debug: Check what's in schema_migrations
	var count int
	db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	logger.Log.Infof("schema_migrations currently has %d records", count)

	if count > 0 {
		rows, _ := db.Query("SELECT version FROM schema_migrations ORDER BY version")
		defer rows.Close()
		for rows.Next() {
			var v string
			rows.Scan(&v)
			logger.Log.Infof("  - %s", v)
		}
	}

	// Get list of migration files
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	sort.Strings(files)
	logger.Log.Infof("Found %d migration files in %s", len(files), migrationsPath)

	// Execute each migration
	for _, file := range files {
		version := filepath.Base(file)

		// Check if migration has already been applied
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		logger.Log.Infof("Checking migration %s: exists=%v", version, exists)

		if exists {
			logger.Log.Infof("Migration %s already applied, skipping", version)
			continue
		}

		// Read and execute migration
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		logger.Log.Infof("Executing migration %s...", version)
		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		logger.Log.Infof("Applied migration: %s", version)
	}

	logger.Log.Info("All migrations completed successfully")
	return nil
}
