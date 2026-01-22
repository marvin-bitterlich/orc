// Package sqlite_test contains integration tests for SQLite repositories.
//
// # Schema Protection
//
// This file is the SINGLE POINT where the database schema is loaded for tests.
// All test setup functions use db.GetSchemaSQL() to ensure tests run against
// the authoritative schema, preventing drift between test and production.
//
// DO NOT hardcode CREATE TABLE statements in test files. `make schema-check`
// will fail if you do. Instead, use setupTestDB() or setupIntegrationDB().
package sqlite_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/db"
)

// setupIntegrationDB creates a shared in-memory database with all tables
// for integration testing scenarios that span multiple repositories.
// Uses the authoritative schema from db.GetSchemaSQL() to prevent drift.
func setupIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Use the authoritative schema from schema.go
	_, err = testDB.Exec(db.GetSchemaSQL())
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		testDB.Close()
	})

	return testDB
}

// seedMission inserts a test commission and returns its ID.
// Note: Function name kept for backwards compatibility with existing tests.
func seedMission(t *testing.T, db *sql.DB, id, title string) string {
	t.Helper()
	_, err := db.Exec("INSERT INTO commissions (id, title, status) VALUES (?, ?, 'active')", id, title)
	if err != nil {
		t.Fatalf("failed to seed commission: %v", err)
	}
	return id
}

// seedGrove inserts a test grove and returns its ID.
func seedGrove(t *testing.T, db *sql.DB, id, missionID, name string) string {
	t.Helper()
	_, err := db.Exec("INSERT INTO groves (id, commission_id, name, status) VALUES (?, ?, ?, 'active')", id, missionID, name)
	if err != nil {
		t.Fatalf("failed to seed grove: %v", err)
	}
	return id
}

// seedTag inserts a test tag and returns its ID.
func seedTag(t *testing.T, db *sql.DB, id, name string) string {
	t.Helper()
	_, err := db.Exec("INSERT INTO tags (id, name) VALUES (?, ?)", id, name)
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}
	return id
}
