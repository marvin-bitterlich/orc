package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var dbInitialized bool

// GetDB returns the database connection, initializing if needed
func GetDB() (*sql.DB, error) {
	if db != nil {
		return db, nil
	}

	// Check for dev/test override via environment variable
	var dbPath, orcDir string
	var err error
	if override := os.Getenv("ORC_DB_PATH"); override != "" {
		dbPath = override
		orcDir = filepath.Dir(override)
	} else {
		var home string
		home, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		orcDir = filepath.Join(home, ".orc")
		dbPath = filepath.Join(orcDir, "orc.db")
	}

	// Ensure .orc directory exists
	if err := os.MkdirAll(orcDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .orc directory: %w", err)
	}

	// Open database connection
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations on first connection (but avoid recursion)
	if !dbInitialized {
		dbInitialized = true
		if err := InitSchema(); err != nil {
			return nil, fmt.Errorf("failed to initialize schema: %w", err)
		}
	}

	return db, nil
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// GetDBPath returns the path to the database file
func GetDBPath() (string, error) {
	// Check for dev/test override via environment variable
	if override := os.Getenv("ORC_DB_PATH"); override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".orc", "orc.db"), nil
}
