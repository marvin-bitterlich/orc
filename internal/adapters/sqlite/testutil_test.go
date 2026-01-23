// Package sqlite_test contains integration tests for SQLite repositories.
//
// # Schema Protection
//
// This file is the SINGLE POINT where the database schema is loaded for tests.
// All test setup functions use db.GetSchemaSQL() to ensure tests run against
// the authoritative schema, preventing drift between test and production.
//
// DO NOT hardcode CREATE TABLE statements in test files. `make schema-check`
// will fail if you do. Instead, use setupTestDB() and the seed* helpers.
package sqlite_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/db"
)

// setupTestDB creates an in-memory database with the authoritative schema.
// This is the single shared test database setup function for all repository tests.
// Uses db.GetSchemaSQL() to prevent test schemas from drifting.
func setupTestDB(t *testing.T) *sql.DB {
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

// seedCommission inserts a test commission and returns its ID.
func seedCommission(t *testing.T, db *sql.DB, id, title string) string {
	t.Helper()
	if id == "" {
		id = "COMM-001"
	}
	if title == "" {
		title = "Test Commission"
	}
	_, err := db.Exec("INSERT INTO commissions (id, title, status) VALUES (?, ?, 'active')", id, title)
	if err != nil {
		t.Fatalf("failed to seed commission: %v", err)
	}
	return id
}

// seedShipment inserts a test shipment and returns its ID.
func seedShipment(t *testing.T, db *sql.DB, id, commissionID, title string) string {
	t.Helper()
	if id == "" {
		id = "SHIP-001"
	}
	if commissionID == "" {
		commissionID = "COMM-001"
	}
	if title == "" {
		title = "Test Shipment"
	}
	_, err := db.Exec("INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, 'active')", id, commissionID, title)
	if err != nil {
		t.Fatalf("failed to seed shipment: %v", err)
	}
	return id
}

// seedWorkbench inserts a test workbench and returns its ID.
func seedWorkbench(t *testing.T, db *sql.DB, id, commissionID, name string) string {
	t.Helper()
	if id == "" {
		id = "GROVE-001"
	}
	if commissionID == "" {
		commissionID = "COMM-001"
	}
	if name == "" {
		name = "test-workbench"
	}
	_, err := db.Exec("INSERT INTO groves (id, commission_id, name, status) VALUES (?, ?, ?, 'active')", id, commissionID, name)
	if err != nil {
		t.Fatalf("failed to seed workbench: %v", err)
	}
	return id
}

// seedTag inserts a test tag and returns its ID.
func seedTag(t *testing.T, db *sql.DB, id, name string) string {
	t.Helper()
	if id == "" {
		id = "TAG-001"
	}
	if name == "" {
		name = "test-tag"
	}
	_, err := db.Exec("INSERT INTO tags (id, name) VALUES (?, ?)", id, name)
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}
	return id
}
