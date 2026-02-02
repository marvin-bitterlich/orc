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
// Schema changes use the Atlas workflow (migrations.go is frozen):
//
//  1. Edit internal/db/schema.sql
//  2. Run: make schema-diff   (preview changes)
//  3. Run: make schema-apply  (apply to local DB)
//  4. Run: make test          (verify alignment)
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
		// Fresh install - check if we have truly OLD schema tables (migrations needed)
		// Note: 'operations' exists in both old and new schemas, so only check for
		// 'expeditions' and 'missions' which were removed in modern schema
		var oldTableCount int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('expeditions', 'missions')").Scan(&oldTableCount)
		if err != nil {
			return err
		}

		if oldTableCount > 0 {
			// Old schema exists - run migrations to upgrade
			return RunMigrations()
		} else {
			// Completely fresh install - create modern schema directly
			// No schema_version table needed; fresh installs are Atlas-managed.
			// The schema.sql uses IF NOT EXISTS so this is idempotent.
			_, err = db.Exec(SchemaSQL)
			return err
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
