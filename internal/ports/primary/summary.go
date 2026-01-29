package primary

import "context"

// SummaryService defines the primary port for summary operations.
// It provides hierarchical views of commissions with role-based filtering.
type SummaryService interface {
	// GetCommissionSummary returns a summary of a commission with role-based filtering.
	GetCommissionSummary(ctx context.Context, req SummaryRequest) (*CommissionSummary, error)
}

// SummaryRequest contains parameters for getting a commission summary.
type SummaryRequest struct {
	CommissionID     string // Required: which commission to summarize
	Role             string // "GOBLIN" or "IMP"
	WorkbenchID      string // For IMP: the workbench making the request
	WorkshopID       string // For GOBLIN: the workshop making the request
	FocusID          string // Currently focused container
	ShowAllShipments bool   // IMP flag: show shipments not assigned to this workbench
	ExpandLibrary    bool   // Show individual tomes in LIBRARY and shipments in SHIPYARD
}

// CommissionSummary represents the hierarchical summary of a commission.
type CommissionSummary struct {
	ID                  string
	Title               string
	Conclaves           []ConclaveSummary
	Library             LibrarySummary
	Shipyard            ShipyardSummary
	HiddenShipmentCount int // Count of shipments hidden from IMP view
}

// ConclaveSummary represents a conclave with its nested tomes and shipments.
type ConclaveSummary struct {
	ID        string
	Title     string
	Status    string
	IsFocused bool
	Pinned    bool
	Tomes     []TomeSummary
	Shipments []ShipmentSummary
}

// TomeSummary represents a tome with its note count.
type TomeSummary struct {
	ID        string
	Title     string
	Status    string
	NoteCount int
	IsFocused bool
	Pinned    bool
}

// ShipmentSummary represents a shipment with task progress.
type ShipmentSummary struct {
	ID         string
	Title      string
	Status     string
	IsFocused  bool
	Pinned     bool
	BenchID    string // Assigned workbench ID (empty if unassigned)
	BenchName  string // Assigned workbench name (for display)
	TasksDone  int
	TasksTotal int
	Tasks      []TaskSummary // Populated only for focused shipment
}

// TaskSummary represents a task in the summary view.
type TaskSummary struct {
	ID     string
	Title  string
	Status string
}

// LibrarySummary represents the Library section with parked tomes.
type LibrarySummary struct {
	TomeCount int
	Tomes     []TomeSummary // Populated when ExpandLibrary is true
}

// ShipyardSummary represents the Shipyard section with parked shipments.
type ShipyardSummary struct {
	ShipmentCount int
	Shipments     []ShipmentSummary // Populated when ExpandLibrary is true
}
