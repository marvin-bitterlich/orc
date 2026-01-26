package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupTaskTestDB creates the test database with required seed data.
func setupTaskTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	seedShipment(t, testDB, "SHIP-001", "COMM-001", "Test Shipment")
	return testDB
}

// createTestTask is a helper that creates a task with a generated ID.
func createTestTask(t *testing.T, repo *sqlite.TaskRepository, ctx context.Context, commissionID, shipmentID, title string) *secondary.TaskRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	task := &secondary.TaskRecord{
		ID:           nextID,
		CommissionID: commissionID,
		ShipmentID:   shipmentID,
		Title:        title,
	}

	err = repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return task
}

func TestTaskRepository_Create(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task := &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		ShipmentID:   "SHIP-001",
		Title:        "Test Task",
		Description:  "A test task description",
		Type:         "implementation",
	}

	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify task was created
	retrieved, err := repo.GetByID(ctx, "TASK-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "ready" {
		t.Errorf("expected status 'ready', got '%s'", retrieved.Status)
	}
	if retrieved.ShipmentID != "SHIP-001" {
		t.Errorf("expected shipment 'SHIP-001', got '%s'", retrieved.ShipmentID)
	}
}

func TestTaskRepository_Create_WithoutShipment(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task := &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Standalone Task",
	}

	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, "TASK-001")
	if retrieved.ShipmentID != "" {
		t.Errorf("expected empty shipment ID, got '%s'", retrieved.ShipmentID)
	}
}

func TestTaskRepository_GetByID(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task := createTestTask(t, repo, ctx, "COMM-001", "SHIP-001", "Test Task")

	retrieved, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", retrieved.Title)
	}
	if retrieved.CommissionID != "COMM-001" {
		t.Errorf("expected commission 'COMM-001', got '%s'", retrieved.CommissionID)
	}
}

func TestTaskRepository_GetByID_NotFound(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "TASK-999")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestTaskRepository_List(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	createTestTask(t, repo, ctx, "COMM-001", "", "Task 1")
	createTestTask(t, repo, ctx, "COMM-001", "", "Task 2")
	createTestTask(t, repo, ctx, "COMM-001", "", "Task 3")

	tasks, err := repo.List(ctx, secondary.TaskFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestTaskRepository_List_FilterByShipment(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	// Add another shipment
	_, _ = db.Exec("INSERT INTO shipments (id, commission_id, title) VALUES ('SHIP-002', 'COMM-001', 'Ship 2')")

	createTestTask(t, repo, ctx, "COMM-001", "SHIP-001", "Task 1")
	createTestTask(t, repo, ctx, "COMM-001", "SHIP-001", "Task 2")
	createTestTask(t, repo, ctx, "COMM-001", "SHIP-002", "Task 3")

	tasks, err := repo.List(ctx, secondary.TaskFilters{ShipmentID: "SHIP-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for SHIP-001, got %d", len(tasks))
	}
}

func TestTaskRepository_List_FilterByStatus(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task1 := createTestTask(t, repo, ctx, "COMM-001", "", "Ready Task")
	createTestTask(t, repo, ctx, "COMM-001", "", "Another Ready Task")

	// Change status of task1
	_ = repo.UpdateStatus(ctx, task1.ID, "in_progress", true, false)

	tasks, err := repo.List(ctx, secondary.TaskFilters{Status: "ready"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 ready task, got %d", len(tasks))
	}
}

func TestTaskRepository_List_FilterByCommission(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Commission 2', 'active')")

	createTestTask(t, repo, ctx, "COMM-001", "", "Task 1")
	createTestTask(t, repo, ctx, "COMM-002", "", "Task 2")

	tasks, err := repo.List(ctx, secondary.TaskFilters{CommissionID: "COMM-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 task for COMM-001, got %d", len(tasks))
	}
}

func TestTaskRepository_Update(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task := createTestTask(t, repo, ctx, "COMM-001", "", "Original Title")

	err := repo.Update(ctx, &secondary.TaskRecord{
		ID:    task.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, task.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestTaskRepository_Update_NotFound(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.TaskRecord{
		ID:    "TASK-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestTaskRepository_Delete(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task := createTestTask(t, repo, ctx, "COMM-001", "", "To Delete")

	err := repo.Delete(ctx, task.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, task.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestTaskRepository_Delete_NotFound(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "TASK-999")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestTaskRepository_Pin_Unpin(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task := createTestTask(t, repo, ctx, "COMM-001", "", "Pin Test")

	// Pin
	err := repo.Pin(ctx, task.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, task.ID)
	if !retrieved.Pinned {
		t.Error("expected task to be pinned")
	}

	// Unpin
	err = repo.Unpin(ctx, task.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, task.ID)
	if retrieved.Pinned {
		t.Error("expected task to be unpinned")
	}
}

func TestTaskRepository_GetNextID(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "TASK-001" {
		t.Errorf("expected TASK-001, got %s", id)
	}

	createTestTask(t, repo, ctx, "COMM-001", "", "Test")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "TASK-002" {
		t.Errorf("expected TASK-002, got %s", id)
	}
}

func TestTaskRepository_UpdateStatus(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task := createTestTask(t, repo, ctx, "COMM-001", "", "Status Test")

	// Update status with claimed timestamp
	err := repo.UpdateStatus(ctx, task.ID, "in_progress", true, false)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, task.ID)
	if retrieved.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", retrieved.Status)
	}
	if retrieved.ClaimedAt == "" {
		t.Error("expected ClaimedAt to be set")
	}
	if retrieved.CompletedAt != "" {
		t.Error("expected CompletedAt to be empty")
	}

	// Update to complete
	err = repo.UpdateStatus(ctx, task.ID, "complete", false, true)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, task.ID)
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
}

func TestTaskRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "TASK-999", "complete", false, true)
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestTaskRepository_Claim(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task := createTestTask(t, repo, ctx, "COMM-001", "", "Claim Test")

	err := repo.Claim(ctx, task.ID, "BENCH-001")
	if err != nil {
		t.Fatalf("Claim failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, task.ID)
	if retrieved.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", retrieved.Status)
	}
	if retrieved.AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected assigned workbench 'BENCH-001', got '%s'", retrieved.AssignedWorkbenchID)
	}
	if retrieved.ClaimedAt == "" {
		t.Error("expected ClaimedAt to be set")
	}
}

func TestTaskRepository_Claim_NotFound(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	err := repo.Claim(ctx, "TASK-999", "BENCH-001")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestTaskRepository_GetByWorkbench(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task1 := createTestTask(t, repo, ctx, "COMM-001", "", "Task 1")
	task2 := createTestTask(t, repo, ctx, "COMM-001", "", "Task 2")
	createTestTask(t, repo, ctx, "COMM-001", "", "Task 3 (unclaimed)")

	_ = repo.Claim(ctx, task1.ID, "BENCH-001")
	_ = repo.Claim(ctx, task2.ID, "BENCH-001")

	tasks, err := repo.GetByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByWorkbench failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for workbench, got %d", len(tasks))
	}
}

func TestTaskRepository_GetByShipment(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	createTestTask(t, repo, ctx, "COMM-001", "SHIP-001", "Task 1")
	createTestTask(t, repo, ctx, "COMM-001", "SHIP-001", "Task 2")
	createTestTask(t, repo, ctx, "COMM-001", "", "Task 3 (no shipment)")

	tasks, err := repo.GetByShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("GetByShipment failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for shipment, got %d", len(tasks))
	}
}

func TestTaskRepository_AssignWorkbenchByShipment(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	task1 := createTestTask(t, repo, ctx, "COMM-001", "SHIP-001", "Task 1")
	task2 := createTestTask(t, repo, ctx, "COMM-001", "SHIP-001", "Task 2")
	task3 := createTestTask(t, repo, ctx, "COMM-001", "", "Task 3 (no shipment)")

	err := repo.AssignWorkbenchByShipment(ctx, "SHIP-001", "BENCH-001")
	if err != nil {
		t.Fatalf("AssignWorkbenchByShipment failed: %v", err)
	}

	// Tasks in shipment should have workbench assigned
	retrieved1, _ := repo.GetByID(ctx, task1.ID)
	if retrieved1.AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected task1 to have workbench assigned")
	}

	retrieved2, _ := repo.GetByID(ctx, task2.ID)
	if retrieved2.AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected task2 to have workbench assigned")
	}

	// Task without shipment should not be affected
	retrieved3, _ := repo.GetByID(ctx, task3.ID)
	if retrieved3.AssignedWorkbenchID != "" {
		t.Errorf("expected task3 to not have workbench assigned")
	}
}

func TestTaskRepository_CommissionExists(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
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

func TestTaskRepository_ShipmentExists(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	exists, err := repo.ShipmentExists(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("ShipmentExists failed: %v", err)
	}
	if !exists {
		t.Error("expected shipment to exist")
	}

	exists, err = repo.ShipmentExists(ctx, "SHIP-999")
	if err != nil {
		t.Fatalf("ShipmentExists failed: %v", err)
	}
	if exists {
		t.Error("expected shipment to not exist")
	}
}

// Tag-related tests

func TestTaskRepository_AddTag_GetTag_RemoveTag(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	// Insert a test tag
	_, _ = db.Exec("INSERT INTO tags (id, name) VALUES ('TAG-001', 'urgent')")

	task := createTestTask(t, repo, ctx, "COMM-001", "", "Tagged Task")

	// Initially no tag
	tag, err := repo.GetTag(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}
	if tag != nil {
		t.Error("expected no tag initially")
	}

	// Add tag
	err = repo.AddTag(ctx, task.ID, "TAG-001")
	if err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	// Get tag
	tag, err = repo.GetTag(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}
	if tag == nil {
		t.Fatal("expected tag to be returned")
	}
	if tag.ID != "TAG-001" {
		t.Errorf("expected tag ID 'TAG-001', got '%s'", tag.ID)
	}
	if tag.Name != "urgent" {
		t.Errorf("expected tag name 'urgent', got '%s'", tag.Name)
	}

	// Remove tag
	err = repo.RemoveTag(ctx, task.ID)
	if err != nil {
		t.Fatalf("RemoveTag failed: %v", err)
	}

	// Tag should be gone
	tag, err = repo.GetTag(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}
	if tag != nil {
		t.Error("expected no tag after removal")
	}
}

func TestTaskRepository_ListByTag(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	// Insert test tags
	_, _ = db.Exec("INSERT INTO tags (id, name) VALUES ('TAG-001', 'urgent')")
	_, _ = db.Exec("INSERT INTO tags (id, name) VALUES ('TAG-002', 'low-priority')")

	task1 := createTestTask(t, repo, ctx, "COMM-001", "", "Task 1")
	task2 := createTestTask(t, repo, ctx, "COMM-001", "", "Task 2")
	createTestTask(t, repo, ctx, "COMM-001", "", "Task 3 (no tag)")

	_ = repo.AddTag(ctx, task1.ID, "TAG-001")
	_ = repo.AddTag(ctx, task2.ID, "TAG-001")

	tasks, err := repo.ListByTag(ctx, "TAG-001")
	if err != nil {
		t.Fatalf("ListByTag failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks with tag, got %d", len(tasks))
	}
}

func TestTaskRepository_GetNextEntityTagID(t *testing.T) {
	db := setupTaskTestDB(t)
	repo := sqlite.NewTaskRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextEntityTagID(ctx)
	if err != nil {
		t.Fatalf("GetNextEntityTagID failed: %v", err)
	}
	if id != "ET-001" {
		t.Errorf("expected ET-001, got %s", id)
	}

	// Insert an entity tag
	_, _ = db.Exec("INSERT INTO entity_tags (id, entity_id, entity_type, tag_id) VALUES ('ET-001', 'TASK-001', 'task', 'TAG-001')")

	id, err = repo.GetNextEntityTagID(ctx)
	if err != nil {
		t.Fatalf("GetNextEntityTagID failed: %v", err)
	}
	if id != "ET-002" {
		t.Errorf("expected ET-002, got %s", id)
	}
}
