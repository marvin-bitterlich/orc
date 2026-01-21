package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func setupQuestionTestDB(t *testing.T) *sql.DB {
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

	// Create investigations table
	_, err = db.Exec(`
		CREATE TABLE investigations (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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

	// Insert test data
	_, _ = db.Exec("INSERT INTO missions (id, title, status) VALUES ('MISSION-001', 'Test Mission', 'active')")
	_, _ = db.Exec("INSERT INTO investigations (id, mission_id, title, status) VALUES ('INV-001', 'MISSION-001', 'Test Investigation', 'active')")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// createTestQuestion is a helper that creates a question with a generated ID.
func createTestQuestion(t *testing.T, repo *sqlite.QuestionRepository, ctx context.Context, missionID, investigationID, title string) *secondary.QuestionRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	question := &secondary.QuestionRecord{
		ID:              nextID,
		MissionID:       missionID,
		InvestigationID: investigationID,
		Title:           title,
	}

	err = repo.Create(ctx, question)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return question
}

func TestQuestionRepository_Create(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	question := &secondary.QuestionRecord{
		ID:              "Q-001",
		MissionID:       "MISSION-001",
		InvestigationID: "INV-001",
		Title:           "Test Question",
		Description:     "A test question description",
	}

	err := repo.Create(ctx, question)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify question was created
	retrieved, err := repo.GetByID(ctx, "Q-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Question" {
		t.Errorf("expected title 'Test Question', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", retrieved.Status)
	}
	if retrieved.InvestigationID != "INV-001" {
		t.Errorf("expected investigation 'INV-001', got '%s'", retrieved.InvestigationID)
	}
}

func TestQuestionRepository_Create_WithoutInvestigation(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	question := &secondary.QuestionRecord{
		ID:        "Q-001",
		MissionID: "MISSION-001",
		Title:     "Standalone Question",
	}

	err := repo.Create(ctx, question)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, "Q-001")
	if retrieved.InvestigationID != "" {
		t.Errorf("expected empty investigation ID, got '%s'", retrieved.InvestigationID)
	}
}

func TestQuestionRepository_GetByID(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	question := createTestQuestion(t, repo, ctx, "MISSION-001", "INV-001", "Test Question")

	retrieved, err := repo.GetByID(ctx, question.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Question" {
		t.Errorf("expected title 'Test Question', got '%s'", retrieved.Title)
	}
	if retrieved.MissionID != "MISSION-001" {
		t.Errorf("expected mission 'MISSION-001', got '%s'", retrieved.MissionID)
	}
}

func TestQuestionRepository_GetByID_NotFound(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "Q-999")
	if err == nil {
		t.Error("expected error for non-existent question")
	}
}

func TestQuestionRepository_List(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	createTestQuestion(t, repo, ctx, "MISSION-001", "", "Question 1")
	createTestQuestion(t, repo, ctx, "MISSION-001", "", "Question 2")
	createTestQuestion(t, repo, ctx, "MISSION-001", "", "Question 3")

	questions, err := repo.List(ctx, secondary.QuestionFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(questions) != 3 {
		t.Errorf("expected 3 questions, got %d", len(questions))
	}
}

func TestQuestionRepository_List_FilterByInvestigation(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	// Add another investigation
	_, _ = db.Exec("INSERT INTO investigations (id, mission_id, title) VALUES ('INV-002', 'MISSION-001', 'Inv 2')")

	createTestQuestion(t, repo, ctx, "MISSION-001", "INV-001", "Question 1")
	createTestQuestion(t, repo, ctx, "MISSION-001", "INV-001", "Question 2")
	createTestQuestion(t, repo, ctx, "MISSION-001", "INV-002", "Question 3")

	questions, err := repo.List(ctx, secondary.QuestionFilters{InvestigationID: "INV-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(questions) != 2 {
		t.Errorf("expected 2 questions for INV-001, got %d", len(questions))
	}
}

func TestQuestionRepository_List_FilterByMission(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	// Add another mission
	_, _ = db.Exec("INSERT INTO missions (id, title, status) VALUES ('MISSION-002', 'Mission 2', 'active')")

	createTestQuestion(t, repo, ctx, "MISSION-001", "", "Question 1")
	createTestQuestion(t, repo, ctx, "MISSION-002", "", "Question 2")

	questions, err := repo.List(ctx, secondary.QuestionFilters{MissionID: "MISSION-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(questions) != 1 {
		t.Errorf("expected 1 question for MISSION-001, got %d", len(questions))
	}
}

func TestQuestionRepository_List_FilterByStatus(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	q1 := createTestQuestion(t, repo, ctx, "MISSION-001", "", "Open Question")
	createTestQuestion(t, repo, ctx, "MISSION-001", "", "Another Open Question")

	// Answer q1
	_ = repo.Answer(ctx, q1.ID, "This is the answer")

	questions, err := repo.List(ctx, secondary.QuestionFilters{Status: "open"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(questions) != 1 {
		t.Errorf("expected 1 open question, got %d", len(questions))
	}
}

func TestQuestionRepository_Update(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	question := createTestQuestion(t, repo, ctx, "MISSION-001", "", "Original Title")

	err := repo.Update(ctx, &secondary.QuestionRecord{
		ID:    question.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, question.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestQuestionRepository_Update_NotFound(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.QuestionRecord{
		ID:    "Q-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent question")
	}
}

func TestQuestionRepository_Delete(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	question := createTestQuestion(t, repo, ctx, "MISSION-001", "", "To Delete")

	err := repo.Delete(ctx, question.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, question.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestQuestionRepository_Delete_NotFound(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "Q-999")
	if err == nil {
		t.Error("expected error for non-existent question")
	}
}

func TestQuestionRepository_Pin_Unpin(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	question := createTestQuestion(t, repo, ctx, "MISSION-001", "", "Pin Test")

	// Pin
	err := repo.Pin(ctx, question.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, question.ID)
	if !retrieved.Pinned {
		t.Error("expected question to be pinned")
	}

	// Unpin
	err = repo.Unpin(ctx, question.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, question.ID)
	if retrieved.Pinned {
		t.Error("expected question to be unpinned")
	}
}

func TestQuestionRepository_Pin_NotFound(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	err := repo.Pin(ctx, "Q-999")
	if err == nil {
		t.Error("expected error for non-existent question")
	}
}

func TestQuestionRepository_GetNextID(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "Q-001" {
		t.Errorf("expected Q-001, got %s", id)
	}

	createTestQuestion(t, repo, ctx, "MISSION-001", "", "Test")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "Q-002" {
		t.Errorf("expected Q-002, got %s", id)
	}
}

func TestQuestionRepository_Answer(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	question := createTestQuestion(t, repo, ctx, "MISSION-001", "", "Question to Answer")

	err := repo.Answer(ctx, question.ID, "This is the answer to the question")
	if err != nil {
		t.Fatalf("Answer failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, question.ID)
	if retrieved.Status != "answered" {
		t.Errorf("expected status 'answered', got '%s'", retrieved.Status)
	}
	if retrieved.Answer != "This is the answer to the question" {
		t.Errorf("expected answer to be set, got '%s'", retrieved.Answer)
	}
	if retrieved.AnsweredAt == "" {
		t.Error("expected AnsweredAt to be set")
	}
}

func TestQuestionRepository_Answer_NotFound(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	err := repo.Answer(ctx, "Q-999", "Answer")
	if err == nil {
		t.Error("expected error for non-existent question")
	}
}

func TestQuestionRepository_MissionExists(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
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

func TestQuestionRepository_InvestigationExists(t *testing.T) {
	db := setupQuestionTestDB(t)
	repo := sqlite.NewQuestionRepository(db)
	ctx := context.Background()

	exists, err := repo.InvestigationExists(ctx, "INV-001")
	if err != nil {
		t.Fatalf("InvestigationExists failed: %v", err)
	}
	if !exists {
		t.Error("expected investigation to exist")
	}

	exists, err = repo.InvestigationExists(ctx, "INV-999")
	if err != nil {
		t.Fatalf("InvestigationExists failed: %v", err)
	}
	if exists {
		t.Error("expected investigation to not exist")
	}
}
