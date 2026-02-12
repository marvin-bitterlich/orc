package db

import (
	"database/sql"
	"fmt"
	"time"
)

// SeedFixtures populates the database with comprehensive development fixtures.
// Uses realistic IDs and data that exercises complex relationships.
func SeedFixtures(database *sql.DB) error {
	now := time.Now().Format(time.RFC3339)

	// Tags
	tags := []struct{ id, name, desc string }{
		{"TAG-001", "urgent", "High priority items"},
		{"TAG-002", "blocked", "Waiting on external dependency"},
		{"TAG-003", "tech-debt", "Technical debt to address"},
	}
	for _, t := range tags {
		if _, err := database.Exec(
			"INSERT INTO tags (id, name, description, created_at) VALUES (?, ?, ?, ?)",
			t.id, t.name, t.desc, now,
		); err != nil {
			return fmt.Errorf("seed tags: %w", err)
		}
	}

	// Repos - Use actual current working directory for orc repo
	repos := []struct{ id, name, path string }{
		{"REPO-001", "orc", "/Users/looneym/wb/orc-45"},
		{"REPO-002", "intercom", "/Users/looneym/src/intercom"},
	}
	for _, r := range repos {
		if _, err := database.Exec(
			"INSERT INTO repos (id, name, local_path, status, created_at) VALUES (?, ?, ?, 'active', ?)",
			r.id, r.name, r.path, now,
		); err != nil {
			return fmt.Errorf("seed repos: %w", err)
		}
	}

	// Factories
	factories := []struct{ id, name string }{
		{"FACT-001", "Main Factory"},
		{"FACT-002", "Test Factory"},
	}
	for _, f := range factories {
		if _, err := database.Exec(
			"INSERT INTO factories (id, name, status, created_at) VALUES (?, ?, 'active', ?)",
			f.id, f.name, now,
		); err != nil {
			return fmt.Errorf("seed factories: %w", err)
		}
	}

	// Workshops
	workshops := []struct{ id, factoryID, name string }{
		{"WORK-001", "FACT-001", "Alpha Workshop"},
		{"WORK-002", "FACT-001", "Beta Workshop"},
		{"WORK-003", "FACT-002", "Test Workshop"},
	}
	for _, w := range workshops {
		if _, err := database.Exec(
			"INSERT INTO workshops (id, factory_id, name, status, created_at) VALUES (?, ?, ?, 'active', ?)",
			w.id, w.factoryID, w.name, now,
		); err != nil {
			return fmt.Errorf("seed workshops: %w", err)
		}
	}

	// Workbenches (path is computed dynamically as ~/wb/{name}, not stored)
	workbenches := []struct{ id, workshopID, name string }{
		{"BENCH-001", "WORK-001", "dev-bench"},
		{"BENCH-002", "WORK-002", "feature-bench"},
	}
	for _, b := range workbenches {
		if _, err := database.Exec(
			"INSERT INTO workbenches (id, workshop_id, name, status, created_at) VALUES (?, ?, ?, 'active', ?)",
			b.id, b.workshopID, b.name, now,
		); err != nil {
			return fmt.Errorf("seed workbenches: %w", err)
		}
	}

	// Commissions
	commissions := []struct{ id, title, status string }{
		{"COMM-001", "ORC Development", "active"},
		{"COMM-002", "Feature Implementation", "active"},
		{"COMM-003", "Old Project", "archived"},
	}
	for _, c := range commissions {
		if _, err := database.Exec(
			"INSERT INTO commissions (id, title, status, created_at) VALUES (?, ?, ?, ?)",
			c.id, c.title, c.status, now,
		); err != nil {
			return fmt.Errorf("seed commissions: %w", err)
		}
	}

	// Shipments (valid statuses: draft, ready, in-progress, closed)
	shipments := []struct{ id, commissionID, title, status string }{
		{"SHIP-001", "COMM-001", "Initial Setup", "closed"},
		{"SHIP-002", "COMM-001", "Core Features", "in-progress"},
		{"SHIP-003", "COMM-001", "Polish & Docs", "draft"},
		{"SHIP-004", "COMM-002", "API Integration", "ready"},
		{"SHIP-005", "COMM-002", "Testing Suite", "draft"},
	}
	for _, s := range shipments {
		if _, err := database.Exec(
			"INSERT INTO shipments (id, commission_id, title, status, created_at) VALUES (?, ?, ?, ?, ?)",
			s.id, s.commissionID, s.title, s.status, now,
		); err != nil {
			return fmt.Errorf("seed shipments: %w", err)
		}
	}

	// Tasks
	tasks := []struct{ id, commissionID, shipmentID, title, status string }{
		{"TASK-001", "COMM-001", "SHIP-001", "Setup repository", "closed"},
		{"TASK-002", "COMM-001", "SHIP-001", "Configure CI", "closed"},
		{"TASK-003", "COMM-001", "SHIP-002", "Implement core logic", "in-progress"},
		{"TASK-004", "COMM-001", "SHIP-002", "Add error handling", "open"},
		{"TASK-005", "COMM-001", "SHIP-002", "Write unit tests", "open"},
		{"TASK-006", "COMM-001", "SHIP-003", "Update README", "open"},
		{"TASK-007", "COMM-002", "SHIP-004", "Design API schema", "open"},
		{"TASK-008", "COMM-002", "SHIP-004", "Implement endpoints", "open"},
		{"TASK-009", "COMM-002", "SHIP-005", "Setup test framework", "open"},
		{"TASK-010", "COMM-002", "SHIP-005", "Write integration tests", "open"},
	}
	for _, t := range tasks {
		if _, err := database.Exec(
			"INSERT INTO tasks (id, commission_id, shipment_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			t.id, t.commissionID, t.shipmentID, t.title, t.status, now,
		); err != nil {
			return fmt.Errorf("seed tasks: %w", err)
		}
	}

	// Tomes
	tomes := []struct{ id, commissionID, title string }{
		{"TOME-001", "COMM-001", "Architecture Notes"},
		{"TOME-002", "COMM-001", "API Exploration"},
	}
	for _, t := range tomes {
		if _, err := database.Exec(
			"INSERT INTO tomes (id, commission_id, title, status, created_at) VALUES (?, ?, ?, 'open', ?)",
			t.id, t.commissionID, t.title, now,
		); err != nil {
			return fmt.Errorf("seed tomes: %w", err)
		}
	}

	// Notes (uses shipment_id, tome_id)
	notesWithTome := []struct{ id, commissionID, tomeID, title, noteType string }{
		{"NOTE-001", "COMM-001", "TOME-001", "Initial thoughts on architecture", "idea"},
		{"NOTE-002", "COMM-001", "TOME-001", "Decision: Use hexagonal architecture", "decision"},
		{"NOTE-003", "COMM-001", "TOME-002", "API versioning approach?", "question"},
	}
	for _, n := range notesWithTome {
		if _, err := database.Exec(
			"INSERT INTO notes (id, commission_id, tome_id, title, type, status, created_at) VALUES (?, ?, ?, ?, ?, 'open', ?)",
			n.id, n.commissionID, n.tomeID, n.title, n.noteType, now,
		); err != nil {
			return fmt.Errorf("seed notes (tome): %w", err)
		}
	}

	notesWithShipment := []struct{ id, commissionID, shipmentID, title, noteType string }{
		{"NOTE-004", "COMM-001", "SHIP-002", "Implementation spec", "spec"},
	}
	for _, n := range notesWithShipment {
		if _, err := database.Exec(
			"INSERT INTO notes (id, commission_id, shipment_id, title, type, status, created_at) VALUES (?, ?, ?, ?, ?, 'open', ?)",
			n.id, n.commissionID, n.shipmentID, n.title, n.noteType, now,
		); err != nil {
			return fmt.Errorf("seed notes (shipment): %w", err)
		}
	}

	return nil
}
