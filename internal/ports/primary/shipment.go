package primary

import "context"

// ShipmentService defines the primary port for shipment operations.
type ShipmentService interface {
	// CreateShipment creates a new shipment for a commission.
	CreateShipment(ctx context.Context, req CreateShipmentRequest) (*CreateShipmentResponse, error)

	// GetShipment retrieves a shipment by ID.
	GetShipment(ctx context.Context, shipmentID string) (*Shipment, error)

	// ListShipments lists shipments with optional filters.
	ListShipments(ctx context.Context, filters ShipmentFilters) ([]*Shipment, error)

	// CompleteShipment marks a shipment as complete.
	// If force is true, completes even if tasks are incomplete.
	CompleteShipment(ctx context.Context, shipmentID string, force bool) error

	// PauseShipment pauses an active shipment.
	PauseShipment(ctx context.Context, shipmentID string) error

	// ResumeShipment resumes a paused shipment.
	ResumeShipment(ctx context.Context, shipmentID string) error

	// UpdateShipment updates a shipment's title and/or description.
	UpdateShipment(ctx context.Context, req UpdateShipmentRequest) error

	// PinShipment pins a shipment to prevent completion.
	PinShipment(ctx context.Context, shipmentID string) error

	// UnpinShipment unpins a shipment.
	UnpinShipment(ctx context.Context, shipmentID string) error

	// AssignShipmentToWorkbench assigns a shipment to a workbench.
	AssignShipmentToWorkbench(ctx context.Context, shipmentID, workbenchID string) error

	// GetShipmentsByWorkbench retrieves shipments assigned to a workbench.
	GetShipmentsByWorkbench(ctx context.Context, workbenchID string) ([]*Shipment, error)

	// GetShipmentTasks retrieves all tasks for a shipment.
	GetShipmentTasks(ctx context.Context, shipmentID string) ([]*Task, error)

	// DeleteShipment deletes a shipment.
	DeleteShipment(ctx context.Context, shipmentID string) error

	// ParkShipment moves a shipment to the commission's Shipyard.
	ParkShipment(ctx context.Context, shipmentID string) error

	// UnparkShipment moves a shipment from Shipyard to a specific Conclave.
	UnparkShipment(ctx context.Context, shipmentID, conclaveID string) error

	// ListShipyardQueue retrieves shipments in the shipyard queue, ordered by priority.
	ListShipyardQueue(ctx context.Context, commissionID string) ([]*ShipyardQueueEntry, error)

	// SetShipmentPriority sets the priority for a shipment in the queue.
	SetShipmentPriority(ctx context.Context, shipmentID string, priority *int) error
}

// CreateShipmentRequest contains parameters for creating a shipment.
type CreateShipmentRequest struct {
	CommissionID  string
	Title         string
	Description   string
	RepoID        string // Optional - link shipment to a repository for branch ownership
	Branch        string // Optional - override auto-generated branch name
	ContainerID   string // Required: CON-xxx or YARD-xxx
	ContainerType string // Required: "conclave" or "shipyard"
}

// CreateShipmentResponse contains the result of creating a shipment.
type CreateShipmentResponse struct {
	ShipmentID string
	Shipment   *Shipment
}

// UpdateShipmentRequest contains parameters for updating a shipment.
type UpdateShipmentRequest struct {
	ShipmentID  string
	Title       string
	Description string
}

// Shipment represents a shipment entity at the port boundary.
type Shipment struct {
	ID                  string
	CommissionID        string
	Title               string
	Description         string
	Status              string
	AssignedWorkbenchID string
	RepoID              string // Linked repository for branch ownership
	Branch              string // Owned branch (e.g., ml/SHIP-001-feature-name)
	Pinned              bool
	ContainerID         string // CON-xxx or YARD-xxx
	ContainerType       string // "conclave" or "shipyard"
	CreatedAt           string
	UpdatedAt           string
	CompletedAt         string
}

// ShipmentFilters contains filter options for listing shipments.
type ShipmentFilters struct {
	CommissionID string
	Status       string
}

// ShipyardQueueEntry represents a shipment in the shipyard queue.
type ShipyardQueueEntry struct {
	ID           string
	CommissionID string
	Title        string
	Priority     *int   // nil = default FIFO, 1 = highest priority
	TaskCount    int    // Total tasks in shipment
	DoneCount    int    // Completed tasks
	CreatedAt    string // When shipment was created
}
