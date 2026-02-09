package primary

import "context"

// HookEventService defines the primary port for hook event operations.
type HookEventService interface {
	// LogHookEvent logs a new hook invocation event.
	LogHookEvent(ctx context.Context, req LogHookEventRequest) (*LogHookEventResponse, error)

	// GetHookEvent retrieves a hook event by ID.
	GetHookEvent(ctx context.Context, eventID string) (*HookEvent, error)

	// ListHookEvents retrieves hook events matching the given filters.
	ListHookEvents(ctx context.Context, filters HookEventFilters) ([]*HookEvent, error)
}

// LogHookEventRequest contains parameters for logging a hook event.
type LogHookEventRequest struct {
	WorkbenchID         string
	HookType            string // 'Stop', 'UserPromptSubmit'
	PayloadJSON         string
	Cwd                 string
	SessionID           string
	ShipmentID          string
	ShipmentStatus      string
	TaskCountIncomplete int    // -1 means null
	Decision            string // 'allow', 'block'
	Reason              string
	DurationMs          int // -1 means null
	Error               string
}

// LogHookEventResponse contains the result of logging a hook event.
type LogHookEventResponse struct {
	EventID string
	Event   *HookEvent
}

// HookEvent represents a hook event entity at the port boundary.
type HookEvent struct {
	ID                  string
	WorkbenchID         string
	HookType            string
	Timestamp           string
	PayloadJSON         string
	Cwd                 string
	SessionID           string
	ShipmentID          string
	ShipmentStatus      string
	TaskCountIncomplete int
	Decision            string
	Reason              string
	DurationMs          int
	Error               string
	CreatedAt           string
}

// HookEventFilters contains filter options for querying hook events.
type HookEventFilters struct {
	WorkbenchID string
	HookType    string
	Limit       int
}

// Hook type constants.
const (
	HookTypeStop             = "Stop"
	HookTypeSubagentStop     = "SubagentStop"
	HookTypeUserPromptSubmit = "UserPromptSubmit"
)

// Hook decision constants.
const (
	HookDecisionAllow = "allow"
	HookDecisionBlock = "block"
)
