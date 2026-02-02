package db

import (
	_ "embed"
)

// SchemaSQL is the complete modern schema for fresh ORC installs.
// This schema reflects the current state after all migrations.
//
// # Schema Drift Protection
//
// This is the SINGLE SOURCE OF TRUTH for the database schema. All tests use
// this schema via GetSchemaSQL(), which provides two layers of protection:
//
//  1. No hardcoded schemas: `make schema-check` fails if any test file contains
//     CREATE TABLE statements. Tests must use db.GetSchemaSQL() instead.
//
//  2. Immediate failure on drift: If repository code references a column that
//     doesn't exist in this schema, tests fail immediately with "no such column".
//     This catches drift at development time, not production.
//
// # Keeping Schema in Sync
//
// IMPORTANT: Keep this in sync with migrations. Use Atlas to verify:
//
//	atlas schema diff --from "sqlite:///$HOME/.orc/orc.db" --to "sqlite:///tmp/fresh.db" --dev-url "sqlite://dev?mode=memory"
//
// When adding new columns or tables:
//  1. Add migration in internal/db/migrations/
//  2. Update schema.sql here
//  3. Run `make test` to verify alignment
//
//go:embed schema.sql
var SchemaSQL string

// InitSchema creates the database schema
func InitSchema() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	// Check if schema_version table exists to determine if this is a fresh install
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&tableCount)
	if err != nil {
		return err
	}

	if tableCount == 0 {
		// Fresh install - check if we have old schema tables (migrations needed)
		var oldTableCount int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('operations', 'expeditions', 'missions')").Scan(&oldTableCount)
		if err != nil {
			return err
		}

		if oldTableCount > 0 {
			// Old schema exists - run migrations to upgrade
			return RunMigrations()
		} else {
			// Completely fresh install - create modern schema directly
			// Also create schema_version at max version to prevent migrations from running
			_, err = db.Exec(SchemaSQL)
			if err != nil {
				return err
			}
			// Mark all migrations as applied for fresh installs
			_, err = db.Exec(`
				CREATE TABLE IF NOT EXISTS schema_version (
					version INTEGER PRIMARY KEY,
					applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
				)
			`)
			if err != nil {
				return err
			}
			// Insert all migration versions as applied
			for i := 1; i <= 47; i++ {
				_, err = db.Exec("INSERT INTO schema_version (version) VALUES (?)", i)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	// schema_version table exists - run any pending migrations
	return RunMigrations()
}

// GetSchemaSQL returns the authoritative schema SQL for use by tests.
// Tests should use this instead of hardcoding their own schema to prevent drift.
func GetSchemaSQL() string {
	return SchemaSQL
}
