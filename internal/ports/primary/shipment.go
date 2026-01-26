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
	CompleteShipment(ctx context.Context, shipmentID string) error

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
}

// CreateShipmentRequest contains parameters for creating a shipment.
type CreateShipmentRequest struct {
	CommissionID string
	Title        string
	Description  string
	RepoID       string // Optional - link shipment to a repository for branch ownership
	Branch       string // Optional - override auto-generated branch name
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
	CreatedAt           string
	UpdatedAt           string
	CompletedAt         string
}

// ShipmentFilters contains filter options for listing shipments.
type ShipmentFilters struct {
	CommissionID string
	Status       string
}
