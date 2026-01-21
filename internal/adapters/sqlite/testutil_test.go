package sqlite_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupIntegrationDB creates a shared in-memory database with all tables
// for integration testing scenarios that span multiple repositories.
func setupIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create commissions table
	_, err = db.Exec(`
		CREATE TABLE commissions (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create commissions table: %v", err)
	}

	// Create groves table
	_, err = db.Exec(`
		CREATE TABLE groves (
			id TEXT PRIMARY KEY,
			commission_id TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create groves table: %v", err)
	}

	// Create shipments table
	_, err = db.Exec(`
		CREATE TABLE shipments (
			id TEXT PRIMARY KEY,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			assigned_workbench_id TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create shipments table: %v", err)
	}

	// Create tasks table
	_, err = db.Exec(`
		CREATE TABLE tasks (
			id TEXT PRIMARY KEY,
			shipment_id TEXT,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT,
			status TEXT NOT NULL DEFAULT 'ready',
			priority TEXT,
			assigned_workbench_id TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			conclave_id TEXT,
			promoted_from_id TEXT,
			promoted_from_type TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create tasks table: %v", err)
	}

	// Create plans table
	_, err = db.Exec(`
		CREATE TABLE plans (
			id TEXT PRIMARY KEY,
			shipment_id TEXT,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'draft',
			content TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			approved_at DATETIME,
			conclave_id TEXT,
			promoted_from_id TEXT,
			promoted_from_type TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create plans table: %v", err)
	}

	// Create investigations table
	_, err = db.Exec(`
		CREATE TABLE investigations (
			id TEXT PRIMARY KEY,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			assigned_workbench_id TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create investigations table: %v", err)
	}

	// Create questions table
	_, err = db.Exec(`
		CREATE TABLE questions (
			id TEXT PRIMARY KEY,
			investigation_id TEXT,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'open',
			answer TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			answered_at DATETIME,
			conclave_id TEXT,
			promoted_from_id TEXT,
			promoted_from_type TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create questions table: %v", err)
	}

	// Create conclaves table
	_, err = db.Exec(`
		CREATE TABLE conclaves (
			id TEXT PRIMARY KEY,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create conclaves table: %v", err)
	}

	// Create tags table
	_, err = db.Exec(`
		CREATE TABLE tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create tags table: %v", err)
	}

	// Create entity_tags junction table
	_, err = db.Exec(`
		CREATE TABLE entity_tags (
			id TEXT PRIMARY KEY,
			entity_id TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create entity_tags table: %v", err)
	}

	// Create operations table
	_, err = db.Exec(`
		CREATE TABLE operations (
			id TEXT PRIMARY KEY,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'ready',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create operations table: %v", err)
	}

	// Create handoffs table
	_, err = db.Exec(`
		CREATE TABLE handoffs (
			id TEXT PRIMARY KEY,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			handoff_note TEXT NOT NULL,
			active_commission_id TEXT,
			active_grove_id TEXT,
			todos_snapshot TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create handoffs table: %v", err)
	}

	// Create messages table
	_, err = db.Exec(`
		CREATE TABLE messages (
			id TEXT PRIMARY KEY,
			sender TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			read INTEGER NOT NULL DEFAULT 0,
			commission_id TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create messages table: %v", err)
	}

	// Create notes table
	_, err = db.Exec(`
		CREATE TABLE notes (
			id TEXT PRIMARY KEY,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			type TEXT,
			status TEXT NOT NULL DEFAULT 'open',
			shipment_id TEXT,
			investigation_id TEXT,
			conclave_id TEXT,
			tome_id TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			closed_at DATETIME,
			promoted_from_id TEXT,
			promoted_from_type TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create notes table: %v", err)
	}

	// Create tomes table
	_, err = db.Exec(`
		CREATE TABLE tomes (
			id TEXT PRIMARY KEY,
			commission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			assigned_workbench_id TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create tomes table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
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
