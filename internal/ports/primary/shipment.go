package primary

import "context"

// ShipmentService defines the primary port for shipment operations.
type ShipmentService interface {
	// CreateShipment creates a new shipment for a mission.
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

	// AssignShipmentToGrove assigns a shipment to a grove.
	AssignShipmentToGrove(ctx context.Context, shipmentID, groveID string) error

	// GetShipmentsByGrove retrieves shipments assigned to a grove.
	GetShipmentsByGrove(ctx context.Context, groveID string) ([]*Shipment, error)

	// GetShipmentTasks retrieves all tasks for a shipment.
	GetShipmentTasks(ctx context.Context, shipmentID string) ([]*Task, error)

	// DeleteShipment deletes a shipment.
	DeleteShipment(ctx context.Context, shipmentID string) error
}

// CreateShipmentRequest contains parameters for creating a shipment.
type CreateShipmentRequest struct {
	MissionID   string
	Title       string
	Description string
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
	ID              string
	MissionID       string
	Title           string
	Description     string
	Status          string
	AssignedGroveID string
	Pinned          bool
	CreatedAt       string
	UpdatedAt       string
	CompletedAt     string
}

// ShipmentFilters contains filter options for listing shipments.
type ShipmentFilters struct {
	MissionID string
	Status    string
}
