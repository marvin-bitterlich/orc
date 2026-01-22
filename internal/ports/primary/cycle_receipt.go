package primary

import "context"

// CycleReceiptService defines the primary port for cycle receipt operations.
type CycleReceiptService interface {
	// CreateCycleReceipt creates a new cycle receipt.
	CreateCycleReceipt(ctx context.Context, req CreateCycleReceiptRequest) (*CreateCycleReceiptResponse, error)

	// GetCycleReceipt retrieves a cycle receipt by ID.
	GetCycleReceipt(ctx context.Context, crecID string) (*CycleReceipt, error)

	// GetCycleReceiptByCWO retrieves a cycle receipt by CWO ID.
	GetCycleReceiptByCWO(ctx context.Context, cwoID string) (*CycleReceipt, error)

	// ListCycleReceipts lists cycle receipts with optional filters.
	ListCycleReceipts(ctx context.Context, filters CycleReceiptFilters) ([]*CycleReceipt, error)

	// UpdateCycleReceipt updates a cycle receipt.
	UpdateCycleReceipt(ctx context.Context, req UpdateCycleReceiptRequest) error

	// DeleteCycleReceipt deletes a cycle receipt.
	DeleteCycleReceipt(ctx context.Context, crecID string) error

	// SubmitCycleReceipt transitions a cycle receipt from draft to submitted.
	SubmitCycleReceipt(ctx context.Context, crecID string) error

	// VerifyCycleReceipt transitions a cycle receipt from submitted to verified.
	VerifyCycleReceipt(ctx context.Context, crecID string) error
}

// CreateCycleReceiptRequest contains parameters for creating a cycle receipt.
type CreateCycleReceiptRequest struct {
	CWOID             string
	DeliveredOutcome  string
	Evidence          string // Optional
	VerificationNotes string // Optional
}

// CreateCycleReceiptResponse contains the result of creating a cycle receipt.
type CreateCycleReceiptResponse struct {
	CycleReceiptID string
	CycleReceipt   *CycleReceipt
}

// UpdateCycleReceiptRequest contains parameters for updating a cycle receipt.
type UpdateCycleReceiptRequest struct {
	CycleReceiptID    string
	DeliveredOutcome  string
	Evidence          string
	VerificationNotes string
}

// CycleReceipt represents a cycle receipt entity at the port boundary.
type CycleReceipt struct {
	ID                string
	CWOID             string
	ShipmentID        string
	DeliveredOutcome  string
	Evidence          string
	VerificationNotes string
	Status            string
	CreatedAt         string
	UpdatedAt         string
}

// CycleReceiptFilters contains filter options for listing cycle receipts.
type CycleReceiptFilters struct {
	CWOID      string
	ShipmentID string
	Status     string
}

// CycleReceipt status constants
const (
	CycleReceiptStatusDraft     = "draft"
	CycleReceiptStatusSubmitted = "submitted"
	CycleReceiptStatusVerified  = "verified"
)
