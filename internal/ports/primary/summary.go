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
	CommissionID string // Required: which commission to summarize
	WorkbenchID  string // The workbench making the request (for context)
	WorkshopID   string // The workshop making the request (for context)
	FocusID      string // Currently focused container (SHIP-xxx or TOME-xxx)
	DebugMode    bool   // Show debug info about what was filtered
}

// CommissionSummary represents the flat summary of a commission.
// Shipments and tomes are listed directly under the commission.
type CommissionSummary struct {
	ID                  string
	Title               string
	IsFocusedCommission bool // true if this is the focused commission
	Shipments           []ShipmentSummary
	Tomes               []TomeSummary
	DebugInfo           *DebugInfo
}

// DebugInfo contains debug messages about filtering decisions.
type DebugInfo struct {
	Messages []string
}

// TomeSummary represents a tome with its note count.
type TomeSummary struct {
	ID        string
	Title     string
	Status    string
	NoteCount int
	IsFocused bool
	Pinned    bool
	Notes     []NoteSummary // Populated when tome is focused
}

// NoteSummary represents a note in the summary view.
type NoteSummary struct {
	ID    string
	Title string
	Type  string // learning, decision, spec, etc.
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
	ID          string
	Title       string
	Status      string
	Plans       []PlanSummary
	Approvals   []ApprovalSummary
	Escalations []EscalationSummary
	Receipts    []ReceiptSummary
}

// PlanSummary represents a plan in the summary view.
type PlanSummary struct {
	ID     string
	Status string // draft, pending_review, approved, escalated, superseded
}

// ApprovalSummary represents an approval in the summary view.
type ApprovalSummary struct {
	ID      string
	Outcome string // approved, escalated
}

// EscalationSummary represents an escalation in the summary view.
type EscalationSummary struct {
	ID            string
	Status        string // pending, resolved, dismissed
	TargetActorID string // GATE-xxx
}

// ReceiptSummary represents a receipt in the summary view.
type ReceiptSummary struct {
	ID     string
	Status string // draft, submitted, verified
}
