package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestHookEventRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewHookEventRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop -> workbench
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, status) VALUES (?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test Workbench", "active")

	t.Run("creates event with all fields", func(t *testing.T) {
		record := &secondary.HookEventRecord{
			ID:                  "HEV-0001",
			WorkbenchID:         "BENCH-001",
			HookType:            "Stop",
			PayloadJSON:         `{"hook_type":"Stop"}`,
			Cwd:                 "/Users/test/project",
			SessionID:           "sess-123",
			ShipmentID:          "SHIP-001",
			ShipmentStatus:      "implementing",
			TaskCountIncomplete: 3,
			Decision:            "block",
			Reason:              "Incomplete tasks",
			DurationMs:          42,
			Error:               "",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "HEV-0001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.WorkbenchID != "BENCH-001" {
			t.Errorf("WorkbenchID = %q, want %q", got.WorkbenchID, "BENCH-001")
		}
		if got.HookType != "Stop" {
			t.Errorf("HookType = %q, want %q", got.HookType, "Stop")
		}
		if got.PayloadJSON != `{"hook_type":"Stop"}` {
			t.Errorf("PayloadJSON = %q, want %q", got.PayloadJSON, `{"hook_type":"Stop"}`)
		}
		if got.Cwd != "/Users/test/project" {
			t.Errorf("Cwd = %q, want %q", got.Cwd, "/Users/test/project")
		}
		if got.SessionID != "sess-123" {
			t.Errorf("SessionID = %q, want %q", got.SessionID, "sess-123")
		}
		if got.ShipmentID != "SHIP-001" {
			t.Errorf("ShipmentID = %q, want %q", got.ShipmentID, "SHIP-001")
		}
		if got.ShipmentStatus != "implementing" {
			t.Errorf("ShipmentStatus = %q, want %q", got.ShipmentStatus, "implementing")
		}
		if got.TaskCountIncomplete != 3 {
			t.Errorf("TaskCountIncomplete = %d, want %d", got.TaskCountIncomplete, 3)
		}
		if got.Decision != "block" {
			t.Errorf("Decision = %q, want %q", got.Decision, "block")
		}
		if got.Reason != "Incomplete tasks" {
			t.Errorf("Reason = %q, want %q", got.Reason, "Incomplete tasks")
		}
		if got.DurationMs != 42 {
			t.Errorf("DurationMs = %d, want %d", got.DurationMs, 42)
		}
	})

	t.Run("creates event with nullable fields null", func(t *testing.T) {
		record := &secondary.HookEventRecord{
			ID:                  "HEV-0002",
			WorkbenchID:         "BENCH-001",
			HookType:            "UserPromptSubmit",
			Decision:            "allow",
			TaskCountIncomplete: -1, // null
			DurationMs:          -1, // null
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "HEV-0002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.PayloadJSON != "" {
			t.Errorf("PayloadJSON = %q, want empty", got.PayloadJSON)
		}
		if got.Cwd != "" {
			t.Errorf("Cwd = %q, want empty", got.Cwd)
		}
		if got.SessionID != "" {
			t.Errorf("SessionID = %q, want empty", got.SessionID)
		}
		if got.ShipmentID != "" {
			t.Errorf("ShipmentID = %q, want empty", got.ShipmentID)
		}
		if got.TaskCountIncomplete != -1 {
			t.Errorf("TaskCountIncomplete = %d, want -1 (null)", got.TaskCountIncomplete)
		}
		if got.Reason != "" {
			t.Errorf("Reason = %q, want empty", got.Reason)
		}
		if got.DurationMs != -1 {
			t.Errorf("DurationMs = %d, want -1 (null)", got.DurationMs)
		}
	})
}

func TestHookEventRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewHookEventRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, status) VALUES (?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test Workbench", "active")

	repo.Create(ctx, &secondary.HookEventRecord{
		ID:                  "HEV-0001",
		WorkbenchID:         "BENCH-001",
		HookType:            "Stop",
		Decision:            "allow",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	})

	t.Run("finds event by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "HEV-0001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "HEV-0001" {
			t.Errorf("ID = %q, want %q", got.ID, "HEV-0001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "HEV-9999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestHookEventRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewHookEventRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, status) VALUES (?, ?, ?, ?)", "BENCH-001", "WORK-001", "Workbench 1", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, status) VALUES (?, ?, ?, ?)", "BENCH-002", "WORK-001", "Workbench 2", "active")

	repo.Create(ctx, &secondary.HookEventRecord{ID: "HEV-0001", WorkbenchID: "BENCH-001", HookType: "Stop", Decision: "block", TaskCountIncomplete: -1, DurationMs: -1})
	repo.Create(ctx, &secondary.HookEventRecord{ID: "HEV-0002", WorkbenchID: "BENCH-001", HookType: "UserPromptSubmit", Decision: "allow", TaskCountIncomplete: -1, DurationMs: -1})
	repo.Create(ctx, &secondary.HookEventRecord{ID: "HEV-0003", WorkbenchID: "BENCH-002", HookType: "Stop", Decision: "allow", TaskCountIncomplete: -1, DurationMs: -1})

	t.Run("lists all events", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.HookEventFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 3 {
			t.Errorf("len = %d, want 3", len(list))
		}
	})

	t.Run("filters by workbench_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.HookEventFilters{WorkbenchID: "BENCH-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by hook_type", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.HookEventFilters{HookType: "Stop"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("applies limit", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.HookEventFilters{Limit: 2})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("combines filters", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.HookEventFilters{WorkbenchID: "BENCH-001", HookType: "Stop"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "HEV-0001" {
			t.Errorf("ID = %q, want %q", list[0].ID, "HEV-0001")
		}
	})
}

func TestHookEventRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewHookEventRepository(db)
	ctx := context.Background()

	t.Run("returns HEV-0001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "HEV-0001" {
			t.Errorf("ID = %q, want %q", id, "HEV-0001")
		}
	})

	t.Run("increments after creating events", func(t *testing.T) {
		// Setup
		db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
		db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
		db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, status) VALUES (?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test Workbench", "active")

		repo.Create(ctx, &secondary.HookEventRecord{
			ID:                  "HEV-0001",
			WorkbenchID:         "BENCH-001",
			HookType:            "Stop",
			Decision:            "allow",
			TaskCountIncomplete: -1,
			DurationMs:          -1,
		})

		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "HEV-0002" {
			t.Errorf("ID = %q, want %q", id, "HEV-0002")
		}
	})
}
