package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port        string
	Environment string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() error {
	// Load .env file if it exists (optional in production)
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return fmt.Errorf("invalid DB_PORT: %w", err)
	}

	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		return fmt.Errorf("invalid JWT_EXPIRY: %w", err)
	}

	AppConfig = &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     port,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "expense_tracker"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiry: jwtExpiry,
		},
		CORS: CORSConfig{
			AllowedOrigins: parseCSVEnv(
				getEnv(
					"ALLOWED_ORIGINS",
					"http://localhost:3000,http://127.0.0.1:3000",
				),
			),
		},
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		parsed, err := parseDatabaseURL(dbURL)
		if err != nil {
			return fmt.Errorf("invalid DATABASE_URL: %w", err)
		}
		AppConfig.Database = parsed
	}

	return nil
}

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseDatabaseURL(rawURL string) (DatabaseConfig, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return DatabaseConfig{}, err
	}

	portStr := u.Port()
	if portStr == "" {
		portStr = "5432"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("invalid port in DATABASE_URL: %w", err)
	}

	password, _ := u.User.Password()
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		dbName = "postgres"
	}

	sslMode := u.Query().Get("sslmode")
	if sslMode == "" {
		sslMode = "require"
	}

	return DatabaseConfig{
		Host:     u.Hostname(),
		Port:     port,
		User:     u.User.Username(),
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}, nil
}

func parseCSVEnv(value string) []string {
	parts := strings.Split(value, ",")
	parsed := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		parsed = append(parsed, trimmed)
	}
	return parsed
}
