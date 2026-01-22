package primary

import "context"

// ReceiptService defines the primary port for receipt operations.
type ReceiptService interface {
	// CreateReceipt creates a new receipt.
	CreateReceipt(ctx context.Context, req CreateReceiptRequest) (*CreateReceiptResponse, error)

	// GetReceipt retrieves a receipt by ID.
	GetReceipt(ctx context.Context, recID string) (*Receipt, error)

	// GetReceiptByShipment retrieves a receipt by shipment ID.
	GetReceiptByShipment(ctx context.Context, shipmentID string) (*Receipt, error)

	// ListReceipts lists receipts with optional filters.
	ListReceipts(ctx context.Context, filters ReceiptFilters) ([]*Receipt, error)

	// UpdateReceipt updates a receipt.
	UpdateReceipt(ctx context.Context, req UpdateReceiptRequest) error

	// DeleteReceipt deletes a receipt.
	DeleteReceipt(ctx context.Context, recID string) error

	// SubmitReceipt transitions a receipt from draft to submitted.
	SubmitReceipt(ctx context.Context, recID string) error

	// VerifyReceipt transitions a receipt from submitted to verified.
	VerifyReceipt(ctx context.Context, recID string) error
}

// CreateReceiptRequest contains parameters for creating a receipt.
type CreateReceiptRequest struct {
	ShipmentID        string
	DeliveredOutcome  string
	Evidence          string // Optional
	VerificationNotes string // Optional
}

// CreateReceiptResponse contains the result of creating a receipt.
type CreateReceiptResponse struct {
	ReceiptID string
	Receipt   *Receipt
}

// UpdateReceiptRequest contains parameters for updating a receipt.
type UpdateReceiptRequest struct {
	ReceiptID         string
	DeliveredOutcome  string
	Evidence          string
	VerificationNotes string
}

// Receipt represents a receipt entity at the port boundary.
type Receipt struct {
	ID                string
	ShipmentID        string
	DeliveredOutcome  string
	Evidence          string
	VerificationNotes string
	Status            string
	CreatedAt         string
	UpdatedAt         string
}

// ReceiptFilters contains filter options for listing receipts.
type ReceiptFilters struct {
	ShipmentID string
	Status     string
}

// Receipt status constants
const (
	ReceiptStatusDraft     = "draft"
	ReceiptStatusSubmitted = "submitted"
	ReceiptStatusVerified  = "verified"
)
