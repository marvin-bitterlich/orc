package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func setupConclaveTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create missions table
	_, err = db.Exec(`
		CREATE TABLE missions (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create missions table: %v", err)
	}

	// Create groves table
	_, err = db.Exec(`
		CREATE TABLE groves (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create groves table: %v", err)
	}

	// Create conclaves table
	_, err = db.Exec(`
		CREATE TABLE conclaves (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			assigned_grove_id TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create conclaves table: %v", err)
	}

	// Create tasks table (for GetTasksByConclave)
	_, err = db.Exec(`
		CREATE TABLE tasks (
			id TEXT PRIMARY KEY,
			shipment_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT,
			status TEXT NOT NULL DEFAULT 'ready',
			priority TEXT,
			assigned_grove_id TEXT,
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

	// Create questions table (for GetQuestionsByConclave)
	_, err = db.Exec(`
		CREATE TABLE questions (
			id TEXT PRIMARY KEY,
			investigation_id TEXT,
			mission_id TEXT NOT NULL,
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

	// Create plans table (for GetPlansByConclave)
	_, err = db.Exec(`
		CREATE TABLE plans (
			id TEXT PRIMARY KEY,
			shipment_id TEXT,
			mission_id TEXT NOT NULL,
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

	// Insert test data
	_, _ = db.Exec("INSERT INTO missions (id, title, status) VALUES ('MISSION-001', 'Test Mission', 'active')")
	_, _ = db.Exec("INSERT INTO groves (id, mission_id, name, status) VALUES ('GROVE-001', 'MISSION-001', 'test-grove', 'active')")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// createTestConclave is a helper that creates a conclave with a generated ID.
func createTestConclave(t *testing.T, repo *sqlite.ConclaveRepository, ctx context.Context, missionID, title, description string) *secondary.ConclaveRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	conclave := &secondary.ConclaveRecord{
		ID:          nextID,
		MissionID:   missionID,
		Title:       title,
		Description: description,
	}

	err = repo.Create(ctx, conclave)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return conclave
}

func TestConclaveRepository_Create(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := &secondary.ConclaveRecord{
		ID:          "CON-001",
		MissionID:   "MISSION-001",
		Title:       "Test Conclave",
		Description: "A test conclave description",
	}

	err := repo.Create(ctx, conclave)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify conclave was created
	retrieved, err := repo.GetByID(ctx, "CON-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Conclave" {
		t.Errorf("expected title 'Test Conclave', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestConclaveRepository_GetByID(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "MISSION-001", "Test Conclave", "Description")

	retrieved, err := repo.GetByID(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Conclave" {
		t.Errorf("expected title 'Test Conclave', got '%s'", retrieved.Title)
	}
	if retrieved.Description != "Description" {
		t.Errorf("expected description 'Description', got '%s'", retrieved.Description)
	}
}

func TestConclaveRepository_GetByID_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "CON-999")
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_List(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	createTestConclave(t, repo, ctx, "MISSION-001", "Conclave 1", "")
	createTestConclave(t, repo, ctx, "MISSION-001", "Conclave 2", "")
	createTestConclave(t, repo, ctx, "MISSION-001", "Conclave 3", "")

	conclaves, err := repo.List(ctx, secondary.ConclaveFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(conclaves) != 3 {
		t.Errorf("expected 3 conclaves, got %d", len(conclaves))
	}
}

func TestConclaveRepository_List_FilterByMission(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	// Add another mission
	_, _ = db.Exec("INSERT INTO missions (id, title, status) VALUES ('MISSION-002', 'Mission 2', 'active')")

	createTestConclave(t, repo, ctx, "MISSION-001", "Conclave 1", "")
	createTestConclave(t, repo, ctx, "MISSION-002", "Conclave 2", "")

	conclaves, err := repo.List(ctx, secondary.ConclaveFilters{MissionID: "MISSION-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(conclaves) != 1 {
		t.Errorf("expected 1 conclave for MISSION-001, got %d", len(conclaves))
	}
}

func TestConclaveRepository_List_FilterByStatus(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	c1 := createTestConclave(t, repo, ctx, "MISSION-001", "Active Conclave", "")
	createTestConclave(t, repo, ctx, "MISSION-001", "Another Active", "")

	// Complete c1
	_ = repo.UpdateStatus(ctx, c1.ID, "complete", true)

	conclaves, err := repo.List(ctx, secondary.ConclaveFilters{Status: "active"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(conclaves) != 1 {
		t.Errorf("expected 1 active conclave, got %d", len(conclaves))
	}
}

func TestConclaveRepository_Update(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "MISSION-001", "Original Title", "")

	err := repo.Update(ctx, &secondary.ConclaveRecord{
		ID:    conclave.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, conclave.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestConclaveRepository_Update_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.ConclaveRecord{
		ID:    "CON-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_Delete(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "MISSION-001", "To Delete", "")

	err := repo.Delete(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, conclave.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestConclaveRepository_Delete_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "CON-999")
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_Pin_Unpin(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "MISSION-001", "Pin Test", "")

	// Pin
	err := repo.Pin(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, conclave.ID)
	if !retrieved.Pinned {
		t.Error("expected conclave to be pinned")
	}

	// Unpin
	err = repo.Unpin(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, conclave.ID)
	if retrieved.Pinned {
		t.Error("expected conclave to be unpinned")
	}
}

func TestConclaveRepository_Pin_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	err := repo.Pin(ctx, "CON-999")
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_GetNextID(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "CON-001" {
		t.Errorf("expected CON-001, got %s", id)
	}

	createTestConclave(t, repo, ctx, "MISSION-001", "Test", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "CON-002" {
		t.Errorf("expected CON-002, got %s", id)
	}
}

func TestConclaveRepository_UpdateStatus(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "MISSION-001", "Status Test", "")

	// Update status without completed timestamp
	err := repo.UpdateStatus(ctx, conclave.ID, "in_progress", false)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, conclave.ID)
	if retrieved.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt != "" {
		t.Error("expected CompletedAt to be empty")
	}

	// Update to complete
	err = repo.UpdateStatus(ctx, conclave.ID, "complete", true)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, conclave.ID)
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
}

func TestConclaveRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "CON-999", "complete", true)
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_GetByGrove(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	// Note: Conclaves don't have AssignGrove, so we need to insert directly
	c1 := createTestConclave(t, repo, ctx, "MISSION-001", "Conclave 1", "")
	c2 := createTestConclave(t, repo, ctx, "MISSION-001", "Conclave 2", "")
	createTestConclave(t, repo, ctx, "MISSION-001", "Conclave 3 (unassigned)", "")

	// Assign groves directly via SQL
	_, _ = db.Exec("UPDATE conclaves SET assigned_grove_id = 'GROVE-001' WHERE id = ?", c1.ID)
	_, _ = db.Exec("UPDATE conclaves SET assigned_grove_id = 'GROVE-001' WHERE id = ?", c2.ID)

	conclaves, err := repo.GetByGrove(ctx, "GROVE-001")
	if err != nil {
		t.Fatalf("GetByGrove failed: %v", err)
	}

	if len(conclaves) != 2 {
		t.Errorf("expected 2 conclaves for grove, got %d", len(conclaves))
	}
}

func TestConclaveRepository_MissionExists(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	exists, err := repo.MissionExists(ctx, "MISSION-001")
	if err != nil {
		t.Fatalf("MissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected mission to exist")
	}

	exists, err = repo.MissionExists(ctx, "MISSION-999")
	if err != nil {
		t.Fatalf("MissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected mission to not exist")
	}
}

// Multi-entity query tests

func TestConclaveRepository_GetTasksByConclave(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "MISSION-001", "Conclave with Tasks", "")

	// Insert tasks for the conclave
	_, _ = db.Exec(`INSERT INTO tasks (id, mission_id, title, status, conclave_id) VALUES ('TASK-001', 'MISSION-001', 'Task 1', 'ready', ?)`, conclave.ID)
	_, _ = db.Exec(`INSERT INTO tasks (id, mission_id, title, status, conclave_id) VALUES ('TASK-002', 'MISSION-001', 'Task 2', 'ready', ?)`, conclave.ID)
	_, _ = db.Exec(`INSERT INTO tasks (id, mission_id, title, status, conclave_id) VALUES ('TASK-003', 'MISSION-001', 'Task 3 (no conclave)', 'ready', NULL)`)

	tasks, err := repo.GetTasksByConclave(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("GetTasksByConclave failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for conclave, got %d", len(tasks))
	}

	// Verify task data
	if tasks[0].Title != "Task 1" {
		t.Errorf("expected title 'Task 1', got '%s'", tasks[0].Title)
	}
}

func TestConclaveRepository_GetQuestionsByConclave(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "MISSION-001", "Conclave with Questions", "")

	// Insert questions for the conclave
	_, _ = db.Exec(`INSERT INTO questions (id, mission_id, title, status, conclave_id) VALUES ('Q-001', 'MISSION-001', 'Question 1', 'open', ?)`, conclave.ID)
	_, _ = db.Exec(`INSERT INTO questions (id, mission_id, title, status, conclave_id) VALUES ('Q-002', 'MISSION-001', 'Question 2', 'open', ?)`, conclave.ID)
	_, _ = db.Exec(`INSERT INTO questions (id, mission_id, title, status, conclave_id) VALUES ('Q-003', 'MISSION-001', 'Question 3 (no conclave)', 'open', NULL)`)

	questions, err := repo.GetQuestionsByConclave(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("GetQuestionsByConclave failed: %v", err)
	}

	if len(questions) != 2 {
		t.Errorf("expected 2 questions for conclave, got %d", len(questions))
	}

	// Verify question data
	if questions[0].Title != "Question 1" {
		t.Errorf("expected title 'Question 1', got '%s'", questions[0].Title)
	}
}

func TestConclaveRepository_GetPlansByConclave(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "MISSION-001", "Conclave with Plans", "")

	// Insert plans for the conclave
	_, _ = db.Exec(`INSERT INTO plans (id, mission_id, title, status, conclave_id) VALUES ('PLAN-001', 'MISSION-001', 'Plan 1', 'draft', ?)`, conclave.ID)
	_, _ = db.Exec(`INSERT INTO plans (id, mission_id, title, status, conclave_id) VALUES ('PLAN-002', 'MISSION-001', 'Plan 2', 'draft', ?)`, conclave.ID)
	_, _ = db.Exec(`INSERT INTO plans (id, mission_id, title, status, conclave_id) VALUES ('PLAN-003', 'MISSION-001', 'Plan 3 (no conclave)', 'draft', NULL)`)

	plans, err := repo.GetPlansByConclave(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("GetPlansByConclave failed: %v", err)
	}

	if len(plans) != 2 {
		t.Errorf("expected 2 plans for conclave, got %d", len(plans))
	}

	// Verify plan data
	if plans[0].Title != "Plan 1" {
		t.Errorf("expected title 'Plan 1', got '%s'", plans[0].Title)
	}
}
