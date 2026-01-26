package primary

import "context"

// TaskService defines the primary port for task operations.
type TaskService interface {
	// CreateTask creates a new task.
	CreateTask(ctx context.Context, req CreateTaskRequest) (*CreateTaskResponse, error)

	// GetTask retrieves a task by ID.
	GetTask(ctx context.Context, taskID string) (*Task, error)

	// ListTasks lists tasks with optional filters.
	ListTasks(ctx context.Context, filters TaskFilters) ([]*Task, error)

	// ClaimTask claims a task for a workbench (sets to in_progress).
	ClaimTask(ctx context.Context, req ClaimTaskRequest) error

	// CompleteTask marks a task as complete.
	CompleteTask(ctx context.Context, taskID string) error

	// PauseTask pauses an in_progress task.
	PauseTask(ctx context.Context, taskID string) error

	// ResumeTask resumes a paused task.
	ResumeTask(ctx context.Context, taskID string) error

	// UpdateTask updates a task's title and/or description.
	UpdateTask(ctx context.Context, req UpdateTaskRequest) error

	// PinTask pins a task to prevent completion.
	PinTask(ctx context.Context, taskID string) error

	// UnpinTask unpins a task.
	UnpinTask(ctx context.Context, taskID string) error

	// GetTasksByWorkbench retrieves tasks assigned to a workbench.
	GetTasksByWorkbench(ctx context.Context, workbenchID string) ([]*Task, error)

	// DeleteTask deletes a task.
	DeleteTask(ctx context.Context, taskID string) error

	// TagTask adds a tag to a task.
	TagTask(ctx context.Context, taskID, tagName string) error

	// UntagTask removes the tag from a task.
	UntagTask(ctx context.Context, taskID string) error

	// ListTasksByTag retrieves tasks with a specific tag.
	ListTasksByTag(ctx context.Context, tagName string) ([]*Task, error)

	// DiscoverTasks finds ready tasks in the current workbench context.
	DiscoverTasks(ctx context.Context, workbenchID string) ([]*Task, error)

	// MoveTask moves a task to a different container.
	MoveTask(ctx context.Context, req MoveTaskRequest) error
}

// CreateTaskRequest contains parameters for creating a task.
type CreateTaskRequest struct {
	ShipmentID      string // Optional
	InvestigationID string // Optional
	CommissionID    string
	Title           string
	Description     string
	Type            string // Optional: research, implementation, fix, documentation, maintenance
}

// CreateTaskResponse contains the result of creating a task.
type CreateTaskResponse struct {
	TaskID string
	Task   *Task
}

// ClaimTaskRequest contains parameters for claiming a task.
type ClaimTaskRequest struct {
	TaskID      string
	WorkbenchID string // Optional, can be derived from context
}

// UpdateTaskRequest contains parameters for updating a task.
type UpdateTaskRequest struct {
	TaskID      string
	Title       string
	Description string
}

// MoveTaskRequest contains parameters for moving a task to a different container.
// Exactly one of the To* fields should be set.
type MoveTaskRequest struct {
	TaskID       string
	ToShipmentID string
	ToTomeID     string
	ToConclaveID string
}

// Task represents a task entity at the port boundary.
type Task struct {
	ID                  string
	ShipmentID          string
	InvestigationID     string
	TomeID              string
	ConclaveID          string
	CommissionID        string
	Title               string
	Description         string
	Type                string
	Status              string
	Priority            string
	AssignedWorkbenchID string
	Pinned              bool
	CreatedAt           string
	UpdatedAt           string
	ClaimedAt           string
	CompletedAt         string
	Tag                 *TaskTag // Populated when retrieving task details
}

// TaskTag represents a tag associated with a task.
type TaskTag struct {
	ID   string
	Name string
}

// TaskFilters contains filter options for listing tasks.
type TaskFilters struct {
	ShipmentID      string
	InvestigationID string
	Status          string
	CommissionID    string
	TagName         string
}
