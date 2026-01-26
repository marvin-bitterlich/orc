package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupNoteTestDB creates the test database with required seed data.
func setupNoteTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	return testDB
}

// createTestNote is a helper that creates a note with a generated ID.
func createTestNote(t *testing.T, repo *sqlite.NoteRepository, ctx context.Context, commissionID, title, content string) *secondary.NoteRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	note := &secondary.NoteRecord{
		ID:           nextID,
		CommissionID: commissionID,
		Title:        title,
		Content:      content,
	}

	err = repo.Create(ctx, note)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return note
}

func TestNoteRepository_Create(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	note := &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Test Note",
		Content:      "This is the note content",
		Type:         "observation",
	}

	err := repo.Create(ctx, note)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify note was created
	retrieved, err := repo.GetByID(ctx, "NOTE-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Note" {
		t.Errorf("expected title 'Test Note', got '%s'", retrieved.Title)
	}
	if retrieved.Content != "This is the note content" {
		t.Errorf("expected correct content, got '%s'", retrieved.Content)
	}
	if retrieved.Type != "observation" {
		t.Errorf("expected type 'observation', got '%s'", retrieved.Type)
	}
}

func TestNoteRepository_Create_WithContainer(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	note := &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Shipment Note",
		ShipmentID:   "SHIP-001",
	}

	err := repo.Create(ctx, note)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, "NOTE-001")
	if retrieved.ShipmentID != "SHIP-001" {
		t.Errorf("expected shipment 'SHIP-001', got '%s'", retrieved.ShipmentID)
	}
}

func TestNoteRepository_GetByID(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	note := createTestNote(t, repo, ctx, "COMM-001", "Test Note", "Content")

	retrieved, err := repo.GetByID(ctx, note.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Note" {
		t.Errorf("expected title 'Test Note', got '%s'", retrieved.Title)
	}
}

func TestNoteRepository_GetByID_NotFound(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "NOTE-999")
	if err == nil {
		t.Error("expected error for non-existent note")
	}
}

func TestNoteRepository_List(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	createTestNote(t, repo, ctx, "COMM-001", "Note 1", "")
	createTestNote(t, repo, ctx, "COMM-001", "Note 2", "")
	createTestNote(t, repo, ctx, "COMM-001", "Note 3", "")

	notes, err := repo.List(ctx, secondary.NoteFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(notes) != 3 {
		t.Errorf("expected 3 notes, got %d", len(notes))
	}
}

func TestNoteRepository_List_FilterByType(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	// Create notes with different types
	n1 := createTestNote(t, repo, ctx, "COMM-001", "Note 1", "")
	n2 := createTestNote(t, repo, ctx, "COMM-001", "Note 2", "")
	createTestNote(t, repo, ctx, "COMM-001", "Note 3", "")

	// Update types directly
	_, _ = db.Exec("UPDATE notes SET type = 'observation' WHERE id = ?", n1.ID)
	_, _ = db.Exec("UPDATE notes SET type = 'observation' WHERE id = ?", n2.ID)
	_, _ = db.Exec("UPDATE notes SET type = 'decision' WHERE id = ?", "NOTE-003")

	notes, err := repo.List(ctx, secondary.NoteFilters{Type: "observation"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("expected 2 observation notes, got %d", len(notes))
	}
}

func TestNoteRepository_List_FilterByCommission(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Commission 2', 'active')")

	createTestNote(t, repo, ctx, "COMM-001", "Note 1", "")
	createTestNote(t, repo, ctx, "COMM-002", "Note 2", "")

	notes, err := repo.List(ctx, secondary.NoteFilters{CommissionID: "COMM-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("expected 1 note for COMM-001, got %d", len(notes))
	}
}

func TestNoteRepository_Update(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	note := createTestNote(t, repo, ctx, "COMM-001", "Original Title", "Original content")

	err := repo.Update(ctx, &secondary.NoteRecord{
		ID:      note.ID,
		Title:   "Updated Title",
		Content: "Updated content",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, note.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
	if retrieved.Content != "Updated content" {
		t.Errorf("expected content 'Updated content', got '%s'", retrieved.Content)
	}
}

func TestNoteRepository_Update_NotFound(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.NoteRecord{
		ID:    "NOTE-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent note")
	}
}

func TestNoteRepository_Delete(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	note := createTestNote(t, repo, ctx, "COMM-001", "To Delete", "")

	err := repo.Delete(ctx, note.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, note.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestNoteRepository_Delete_NotFound(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "NOTE-999")
	if err == nil {
		t.Error("expected error for non-existent note")
	}
}

func TestNoteRepository_Pin_Unpin(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	note := createTestNote(t, repo, ctx, "COMM-001", "Pin Test", "")

	// Pin
	err := repo.Pin(ctx, note.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, note.ID)
	if !retrieved.Pinned {
		t.Error("expected note to be pinned")
	}

	// Unpin
	err = repo.Unpin(ctx, note.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, note.ID)
	if retrieved.Pinned {
		t.Error("expected note to be unpinned")
	}
}

func TestNoteRepository_Pin_NotFound(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	err := repo.Pin(ctx, "NOTE-999")
	if err == nil {
		t.Error("expected error for non-existent note")
	}
}

func TestNoteRepository_GetNextID(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "NOTE-001" {
		t.Errorf("expected NOTE-001, got %s", id)
	}

	createTestNote(t, repo, ctx, "COMM-001", "Test", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "NOTE-002" {
		t.Errorf("expected NOTE-002, got %s", id)
	}
}

// Container query tests

func TestNoteRepository_GetByContainer_Shipment(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	// Create notes for different containers
	_, _ = db.Exec(`INSERT INTO notes (id, commission_id, title, shipment_id) VALUES ('NOTE-001', 'COMM-001', 'Ship Note 1', 'SHIP-001')`)
	_, _ = db.Exec(`INSERT INTO notes (id, commission_id, title, shipment_id) VALUES ('NOTE-002', 'COMM-001', 'Ship Note 2', 'SHIP-001')`)
	_, _ = db.Exec(`INSERT INTO notes (id, commission_id, title, investigation_id) VALUES ('NOTE-003', 'COMM-001', 'Inv Note', 'INV-001')`)

	notes, err := repo.GetByContainer(ctx, "shipment", "SHIP-001")
	if err != nil {
		t.Fatalf("GetByContainer failed: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("expected 2 notes for shipment, got %d", len(notes))
	}
}

func TestNoteRepository_GetByContainer_Investigation(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	_, _ = db.Exec(`INSERT INTO notes (id, commission_id, title, investigation_id) VALUES ('NOTE-001', 'COMM-001', 'Inv Note 1', 'INV-001')`)
	_, _ = db.Exec(`INSERT INTO notes (id, commission_id, title, investigation_id) VALUES ('NOTE-002', 'COMM-001', 'Inv Note 2', 'INV-001')`)
	_, _ = db.Exec(`INSERT INTO notes (id, commission_id, title, investigation_id) VALUES ('NOTE-003', 'COMM-001', 'Inv Note 3', 'INV-002')`)

	notes, err := repo.GetByContainer(ctx, "investigation", "INV-001")
	if err != nil {
		t.Fatalf("GetByContainer failed: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("expected 2 notes for investigation, got %d", len(notes))
	}
}

func TestNoteRepository_GetByContainer_Conclave(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	_, _ = db.Exec(`INSERT INTO notes (id, commission_id, title, conclave_id) VALUES ('NOTE-001', 'COMM-001', 'Conclave Note', 'CON-001')`)

	notes, err := repo.GetByContainer(ctx, "conclave", "CON-001")
	if err != nil {
		t.Fatalf("GetByContainer failed: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("expected 1 note for conclave, got %d", len(notes))
	}
}

func TestNoteRepository_GetByContainer_Tome(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	_, _ = db.Exec(`INSERT INTO notes (id, commission_id, title, tome_id) VALUES ('NOTE-001', 'COMM-001', 'Tome Note', 'TOME-001')`)

	notes, err := repo.GetByContainer(ctx, "tome", "TOME-001")
	if err != nil {
		t.Fatalf("GetByContainer failed: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("expected 1 note for tome, got %d", len(notes))
	}
}

func TestNoteRepository_GetByContainer_UnknownType(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	_, err := repo.GetByContainer(ctx, "unknown", "ID-001")
	if err == nil {
		t.Error("expected error for unknown container type")
	}
}

func TestNoteRepository_CommissionExists(t *testing.T) {
	db := setupNoteTestDB(t)
	repo := sqlite.NewNoteRepository(db)
	ctx := context.Background()

	exists, err := repo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected commission to exist")
	}

	exists, err = repo.CommissionExists(ctx, "COMM-999")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected commission to not exist")
	}
}
