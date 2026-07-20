package config

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	// Strict requirement: using pure database/sql with the lib/pq driver
	_ "github.com/lib/pq"
)

// Config holds all configuration settings parsed from environment variables.
type Config struct {
	AppPort int
	DBConn  string
}

// LoadConfig parses environment variables and returns a validated Config struct.
// It falls back to safe development defaults if variables are missing.
func LoadConfig() (*Config, error) {
	portStr := getEnv("APP_PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT value %q: %w", portStr, err)
	}

	// Constructing standard PostgreSQL connection string format
	// e.g., "host=localhost port=5432 user=postgres password=secret dbname=library sslmode=disable"
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "secret")
	dbName := getEnv("DB_NAME", "library_db")
	dbSSL := getEnv("DB_SSLMODE", "disable")

	dbConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, dbSSL,
	)

	return &Config{
		AppPort: port,
		DBConn:  dbConnStr,
	}, nil
}

// InitDB sets up a heavily optimized sql.DB connection pool for PostgreSQL.
func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Production connection pool tuning
	// Sets the maximum number of open connections to the database.
	db.SetMaxOpenConns(25)
	// Sets the maximum number of connections in the idle connection pool.
	db.SetMaxIdleConns(25)
	// Sets the maximum amount of time a connection may be reused.
	// Keeps it lower than database-side timeouts (usually 1 hour) to avoid stale links.
	db.SetConnMaxLifetime(15 * time.Minute)
	// Sets the maximum amount of time a connection may be idle before being closed.
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Explicitly verify the database is reachable before returning the pool
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil
}

// getEnv checks for an environment variable or returns a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
