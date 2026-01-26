package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// TaskServiceImpl implements the TaskService interface.
type TaskServiceImpl struct {
	taskRepo secondary.TaskRepository
	tagRepo  secondary.TagRepository
}

// NewTaskService creates a new TaskService with injected dependencies.
func NewTaskService(
	taskRepo secondary.TaskRepository,
	tagRepo secondary.TagRepository,
) *TaskServiceImpl {
	return &TaskServiceImpl{
		taskRepo: taskRepo,
		tagRepo:  tagRepo,
	}
}

// CreateTask creates a new task.
func (s *TaskServiceImpl) CreateTask(ctx context.Context, req primary.CreateTaskRequest) (*primary.CreateTaskResponse, error) {
	// Validate commission exists
	exists, err := s.taskRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	// Validate shipment exists if provided
	if req.ShipmentID != "" {
		exists, err := s.taskRepo.ShipmentExists(ctx, req.ShipmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate shipment: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("shipment %s not found", req.ShipmentID)
		}
	}

	// Get next ID
	nextID, err := s.taskRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate task ID: %w", err)
	}

	// Create record
	record := &secondary.TaskRecord{
		ID:              nextID,
		ShipmentID:      req.ShipmentID,
		InvestigationID: req.InvestigationID,
		CommissionID:    req.CommissionID,
		Title:           req.Title,
		Description:     req.Description,
		Type:            req.Type,
		Status:          "ready",
	}

	if err := s.taskRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Fetch created task
	created, err := s.taskRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created task: %w", err)
	}

	return &primary.CreateTaskResponse{
		TaskID: created.ID,
		Task:   recordToTask(created),
	}, nil
}

// GetTask retrieves a task by ID.
func (s *TaskServiceImpl) GetTask(ctx context.Context, taskID string) (*primary.Task, error) {
	record, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task := recordToTask(record)

	// Optionally load tag
	tag, err := s.taskRepo.GetTag(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task tag: %w", err)
	}
	if tag != nil {
		task.Tag = &primary.TaskTag{
			ID:   tag.ID,
			Name: tag.Name,
		}
	}

	return task, nil
}

// ListTasks lists tasks with optional filters.
func (s *TaskServiceImpl) ListTasks(ctx context.Context, filters primary.TaskFilters) ([]*primary.Task, error) {
	records, err := s.taskRepo.List(ctx, secondary.TaskFilters{
		ShipmentID:      filters.ShipmentID,
		InvestigationID: filters.InvestigationID,
		Status:          filters.Status,
		CommissionID:    filters.CommissionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasks := make([]*primary.Task, len(records))
	for i, r := range records {
		tasks[i] = recordToTask(r)
	}
	return tasks, nil
}

// ClaimTask claims a task for a workbench.
func (s *TaskServiceImpl) ClaimTask(ctx context.Context, req primary.ClaimTaskRequest) error {
	// Verify task exists
	_, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return err
	}

	return s.taskRepo.Claim(ctx, req.TaskID, req.WorkbenchID)
}

// CompleteTask marks a task as complete.
func (s *TaskServiceImpl) CompleteTask(ctx context.Context, taskID string) error {
	record, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Guard: cannot complete pinned task
	if record.Pinned {
		return fmt.Errorf("cannot complete pinned task %s. Unpin first with: orc task unpin %s", taskID, taskID)
	}

	return s.taskRepo.UpdateStatus(ctx, taskID, "complete", false, true)
}

// PauseTask pauses an in_progress task.
func (s *TaskServiceImpl) PauseTask(ctx context.Context, taskID string) error {
	record, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Guard: can only pause in_progress tasks
	if record.Status != "in_progress" {
		return fmt.Errorf("can only pause in_progress tasks (current status: %s)", record.Status)
	}

	return s.taskRepo.UpdateStatus(ctx, taskID, "paused", false, false)
}

// ResumeTask resumes a paused task.
func (s *TaskServiceImpl) ResumeTask(ctx context.Context, taskID string) error {
	record, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Guard: can only resume paused tasks
	if record.Status != "paused" {
		return fmt.Errorf("can only resume paused tasks (current status: %s)", record.Status)
	}

	return s.taskRepo.UpdateStatus(ctx, taskID, "in_progress", false, false)
}

// UpdateTask updates a task's title and/or description.
func (s *TaskServiceImpl) UpdateTask(ctx context.Context, req primary.UpdateTaskRequest) error {
	record := &secondary.TaskRecord{
		ID:          req.TaskID,
		Title:       req.Title,
		Description: req.Description,
	}
	return s.taskRepo.Update(ctx, record)
}

// PinTask pins a task.
func (s *TaskServiceImpl) PinTask(ctx context.Context, taskID string) error {
	return s.taskRepo.Pin(ctx, taskID)
}

// UnpinTask unpins a task.
func (s *TaskServiceImpl) UnpinTask(ctx context.Context, taskID string) error {
	return s.taskRepo.Unpin(ctx, taskID)
}

// GetTasksByWorkbench retrieves tasks assigned to a workbench.
func (s *TaskServiceImpl) GetTasksByWorkbench(ctx context.Context, workbenchID string) ([]*primary.Task, error) {
	records, err := s.taskRepo.GetByWorkbench(ctx, workbenchID)
	if err != nil {
		return nil, err
	}

	tasks := make([]*primary.Task, len(records))
	for i, r := range records {
		tasks[i] = recordToTask(r)
	}
	return tasks, nil
}

// DeleteTask deletes a task.
func (s *TaskServiceImpl) DeleteTask(ctx context.Context, taskID string) error {
	return s.taskRepo.Delete(ctx, taskID)
}

// TagTask adds a tag to a task.
func (s *TaskServiceImpl) TagTask(ctx context.Context, taskID, tagName string) error {
	// Verify task exists
	_, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Get tag by name
	tag, err := s.tagRepo.GetByName(ctx, tagName)
	if err != nil {
		return fmt.Errorf("tag '%s' not found", tagName)
	}

	// Check if task already has a tag
	existingTag, err := s.taskRepo.GetTag(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to check existing tag: %w", err)
	}
	if existingTag != nil {
		return fmt.Errorf("task %s already has tag '%s' (one tag per task limit)\nRemove existing tag first with: orc task untag %s", taskID, existingTag.Name, taskID)
	}

	return s.taskRepo.AddTag(ctx, taskID, tag.ID)
}

// UntagTask removes the tag from a task.
func (s *TaskServiceImpl) UntagTask(ctx context.Context, taskID string) error {
	// Verify task exists
	_, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Check if task has a tag
	tag, err := s.taskRepo.GetTag(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task tag: %w", err)
	}
	if tag == nil {
		return fmt.Errorf("task %s has no tag assigned", taskID)
	}

	return s.taskRepo.RemoveTag(ctx, taskID)
}

// ListTasksByTag retrieves tasks with a specific tag.
func (s *TaskServiceImpl) ListTasksByTag(ctx context.Context, tagName string) ([]*primary.Task, error) {
	// Get tag by name
	tag, err := s.tagRepo.GetByName(ctx, tagName)
	if err != nil {
		return nil, fmt.Errorf("tag '%s' not found", tagName)
	}

	records, err := s.taskRepo.ListByTag(ctx, tag.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks by tag: %w", err)
	}

	tasks := make([]*primary.Task, len(records))
	for i, r := range records {
		tasks[i] = recordToTask(r)
	}
	return tasks, nil
}

// DiscoverTasks finds ready tasks in the current workbench context.
func (s *TaskServiceImpl) DiscoverTasks(ctx context.Context, workbenchID string) ([]*primary.Task, error) {
	records, err := s.taskRepo.GetByWorkbench(ctx, workbenchID)
	if err != nil {
		return nil, err
	}

	// Filter to ready tasks
	var readyTasks []*primary.Task
	for _, r := range records {
		if r.Status == "ready" {
			readyTasks = append(readyTasks, recordToTask(r))
		}
	}
	return readyTasks, nil
}

// MoveTask moves a task to a different container.
func (s *TaskServiceImpl) MoveTask(ctx context.Context, req primary.MoveTaskRequest) error {
	// Verify task exists
	_, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return err
	}

	// Count how many targets are specified - exactly one required
	targetCount := 0
	if req.ToShipmentID != "" {
		targetCount++
	}
	if req.ToTomeID != "" {
		targetCount++
	}
	if req.ToConclaveID != "" {
		targetCount++
	}

	if targetCount == 0 {
		return fmt.Errorf("must specify exactly one target container (--to-shipment, --to-tome, or --to-conclave)")
	}
	if targetCount > 1 {
		return fmt.Errorf("cannot specify multiple target containers")
	}

	// Validate target container exists and build update record
	record := &secondary.TaskRecord{ID: req.TaskID}

	if req.ToShipmentID != "" {
		exists, err := s.taskRepo.ShipmentExists(ctx, req.ToShipmentID)
		if err != nil {
			return fmt.Errorf("failed to validate shipment: %w", err)
		}
		if !exists {
			return fmt.Errorf("shipment %s not found", req.ToShipmentID)
		}
		record.ShipmentID = req.ToShipmentID
	}

	if req.ToTomeID != "" {
		exists, err := s.taskRepo.TomeExists(ctx, req.ToTomeID)
		if err != nil {
			return fmt.Errorf("failed to validate tome: %w", err)
		}
		if !exists {
			return fmt.Errorf("tome %s not found", req.ToTomeID)
		}
		record.TomeID = req.ToTomeID
	}

	if req.ToConclaveID != "" {
		exists, err := s.taskRepo.ConclaveExists(ctx, req.ToConclaveID)
		if err != nil {
			return fmt.Errorf("failed to validate conclave: %w", err)
		}
		if !exists {
			return fmt.Errorf("conclave %s not found", req.ToConclaveID)
		}
		record.ConclaveID = req.ToConclaveID
	}

	return s.taskRepo.Update(ctx, record)
}

// Ensure TaskServiceImpl implements the interface
var _ primary.TaskService = (*TaskServiceImpl)(nil)
