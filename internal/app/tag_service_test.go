package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// mockTagRepository implements secondary.TagRepository for testing.
type mockTagRepository struct {
	tags       map[string]*secondary.TagRecord
	entityTags map[string]*secondary.TagRecord // entityID -> tag
	createErr  error
	getErr     error
	deleteErr  error
	listErr    error
}

func newMockTagRepository() *mockTagRepository {
	return &mockTagRepository{
		tags:       make(map[string]*secondary.TagRecord),
		entityTags: make(map[string]*secondary.TagRecord),
	}
}

func (m *mockTagRepository) Create(ctx context.Context, tag *secondary.TagRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.tags[tag.ID] = tag
	return nil
}

func (m *mockTagRepository) GetByID(ctx context.Context, id string) (*secondary.TagRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if tag, ok := m.tags[id]; ok {
		return tag, nil
	}
	return nil, errors.New("tag not found")
}

func (m *mockTagRepository) GetByName(ctx context.Context, name string) (*secondary.TagRecord, error) {
	for _, tag := range m.tags {
		if tag.Name == name {
			return tag, nil
		}
	}
	return nil, errors.New("tag not found")
}

func (m *mockTagRepository) List(ctx context.Context) ([]*secondary.TagRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.TagRecord
	for _, tag := range m.tags {
		result = append(result, tag)
	}
	return result, nil
}

func (m *mockTagRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.tags, id)
	return nil
}

func (m *mockTagRepository) GetNextID(ctx context.Context) (string, error) {
	return "TAG-001", nil
}

func (m *mockTagRepository) GetEntityTag(ctx context.Context, entityID, entityType string) (*secondary.TagRecord, error) {
	key := entityType + ":" + entityID
	if tag, ok := m.entityTags[key]; ok {
		return tag, nil
	}
	return nil, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestTagService() (*TagServiceImpl, *mockTagRepository) {
	tagRepo := newMockTagRepository()
	service := NewTagService(tagRepo)
	return service, tagRepo
}

// ============================================================================
// CreateTag Tests
// ============================================================================

func TestCreateTag_Success(t *testing.T) {
	service, _ := newTestTagService()
	ctx := context.Background()

	resp, err := service.CreateTag(ctx, primary.CreateTagRequest{
		Name:        "urgent",
		Description: "High priority tasks",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.TagID == "" {
		t.Error("expected tag ID to be set")
	}
	if resp.Tag.Name != "urgent" {
		t.Errorf("expected name 'urgent', got '%s'", resp.Tag.Name)
	}
	if resp.Tag.Description != "High priority tasks" {
		t.Errorf("expected description 'High priority tasks', got '%s'", resp.Tag.Description)
	}
}

func TestCreateTag_MinimalFields(t *testing.T) {
	service, _ := newTestTagService()
	ctx := context.Background()

	resp, err := service.CreateTag(ctx, primary.CreateTagRequest{
		Name: "bug",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Tag.Name != "bug" {
		t.Errorf("expected name 'bug', got '%s'", resp.Tag.Name)
	}
}

// ============================================================================
// GetTag Tests
// ============================================================================

func TestGetTag_Found(t *testing.T) {
	service, tagRepo := newTestTagService()
	ctx := context.Background()

	tagRepo.tags["TAG-001"] = &secondary.TagRecord{
		ID:          "TAG-001",
		Name:        "urgent",
		Description: "High priority",
	}

	tag, err := service.GetTag(ctx, "TAG-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tag.Name != "urgent" {
		t.Errorf("expected name 'urgent', got '%s'", tag.Name)
	}
}

func TestGetTag_NotFound(t *testing.T) {
	service, _ := newTestTagService()
	ctx := context.Background()

	_, err := service.GetTag(ctx, "TAG-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent tag, got nil")
	}
}

// ============================================================================
// GetTagByName Tests
// ============================================================================

func TestGetTagByName_Found(t *testing.T) {
	service, tagRepo := newTestTagService()
	ctx := context.Background()

	tagRepo.tags["TAG-001"] = &secondary.TagRecord{
		ID:          "TAG-001",
		Name:        "urgent",
		Description: "High priority",
	}

	tag, err := service.GetTagByName(ctx, "urgent")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tag.ID != "TAG-001" {
		t.Errorf("expected ID 'TAG-001', got '%s'", tag.ID)
	}
}

func TestGetTagByName_NotFound(t *testing.T) {
	service, _ := newTestTagService()
	ctx := context.Background()

	_, err := service.GetTagByName(ctx, "nonexistent")

	if err == nil {
		t.Fatal("expected error for non-existent tag name, got nil")
	}
}

// ============================================================================
// ListTags Tests
// ============================================================================

func TestListTags_Success(t *testing.T) {
	service, tagRepo := newTestTagService()
	ctx := context.Background()

	tagRepo.tags["TAG-001"] = &secondary.TagRecord{
		ID:   "TAG-001",
		Name: "urgent",
	}
	tagRepo.tags["TAG-002"] = &secondary.TagRecord{
		ID:   "TAG-002",
		Name: "bug",
	}
	tagRepo.tags["TAG-003"] = &secondary.TagRecord{
		ID:   "TAG-003",
		Name: "feature",
	}

	tags, err := service.ListTags(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
}

func TestListTags_Empty(t *testing.T) {
	service, _ := newTestTagService()
	ctx := context.Background()

	tags, err := service.ListTags(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tags == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

// ============================================================================
// DeleteTag Tests
// ============================================================================

func TestDeleteTag_Success(t *testing.T) {
	service, tagRepo := newTestTagService()
	ctx := context.Background()

	tagRepo.tags["TAG-001"] = &secondary.TagRecord{
		ID:   "TAG-001",
		Name: "urgent",
	}

	err := service.DeleteTag(ctx, "TAG-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := tagRepo.tags["TAG-001"]; exists {
		t.Error("expected tag to be deleted")
	}
}

// ============================================================================
// GetEntityTag Tests
// ============================================================================

func TestGetEntityTag_Found(t *testing.T) {
	service, tagRepo := newTestTagService()
	ctx := context.Background()

	tagRepo.tags["TAG-001"] = &secondary.TagRecord{
		ID:   "TAG-001",
		Name: "urgent",
	}
	tagRepo.entityTags["task:TASK-001"] = tagRepo.tags["TAG-001"]

	tag, err := service.GetEntityTag(ctx, "TASK-001", "task")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tag == nil {
		t.Fatal("expected tag, got nil")
	}
	if tag.Name != "urgent" {
		t.Errorf("expected name 'urgent', got '%s'", tag.Name)
	}
}

func TestGetEntityTag_NotFound(t *testing.T) {
	service, _ := newTestTagService()
	ctx := context.Background()

	tag, err := service.GetEntityTag(ctx, "TASK-001", "task")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tag != nil {
		t.Error("expected nil tag for entity without tag")
	}
}
