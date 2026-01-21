package primary

import "context"

// PRService defines the primary port for pull request operations.
type PRService interface {
	// CreatePR creates a new pull request linked to a shipment.
	CreatePR(ctx context.Context, req CreatePRRequest) (*CreatePRResponse, error)

	// GetPR retrieves a pull request by ID.
	GetPR(ctx context.Context, prID string) (*PR, error)

	// GetPRByShipment retrieves a pull request by its associated shipment.
	GetPRByShipment(ctx context.Context, shipmentID string) (*PR, error)

	// ListPRs lists pull requests with optional filters.
	ListPRs(ctx context.Context, filters PRFilters) ([]*PR, error)

	// UpdatePR updates a pull request's metadata.
	UpdatePR(ctx context.Context, req UpdatePRRequest) error

	// OpenPR opens a draft PR for review.
	OpenPR(ctx context.Context, prID string) error

	// ApprovePR marks a PR as approved.
	ApprovePR(ctx context.Context, prID string) error

	// MergePR merges a PR (cascades to complete the shipment).
	MergePR(ctx context.Context, prID string) error

	// ClosePR closes a PR without merging.
	ClosePR(ctx context.Context, prID string) error

	// LinkPR links an existing external PR to a shipment.
	LinkPR(ctx context.Context, shipmentID, url string, number int) (*PR, error)
}

// CreatePRRequest contains parameters for creating a pull request.
type CreatePRRequest struct {
	ShipmentID   string
	RepoID       string
	Title        string
	Description  string
	Branch       string
	TargetBranch string
	Draft        bool
	URL          string // For linking existing PRs
	Number       int    // GitHub PR number
}

// CreatePRResponse contains the result of creating a pull request.
type CreatePRResponse struct {
	PRID string
	PR   *PR
}

// UpdatePRRequest contains parameters for updating a pull request.
type UpdatePRRequest struct {
	PRID        string
	Title       string
	Description string
	URL         string
	Number      int
}

// PR represents a pull request entity at the port boundary.
type PR struct {
	ID           string
	ShipmentID   string
	RepoID       string
	CommissionID string
	Number       int
	Title        string
	Description  string
	Branch       string
	TargetBranch string
	URL          string
	Status       string
	CreatedAt    string
	UpdatedAt    string
	MergedAt     string
	ClosedAt     string
}

// PRFilters contains filter options for listing pull requests.
type PRFilters struct {
	ShipmentID   string
	RepoID       string
	CommissionID string
	Status       string
}

// PR status constants
const (
	PRStatusDraft    = "draft"
	PRStatusOpen     = "open"
	PRStatusApproved = "approved"
	PRStatusMerged   = "merged"
	PRStatusClosed   = "closed"
)
