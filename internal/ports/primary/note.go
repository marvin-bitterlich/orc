package primary

import "context"

// NoteService defines the primary port for note operations.
type NoteService interface {
	// CreateNote creates a new note.
	CreateNote(ctx context.Context, req CreateNoteRequest) (*CreateNoteResponse, error)

	// GetNote retrieves a note by ID.
	GetNote(ctx context.Context, noteID string) (*Note, error)

	// ListNotes lists notes with optional filters.
	ListNotes(ctx context.Context, filters NoteFilters) ([]*Note, error)

	// UpdateNote updates a note's title and/or content.
	UpdateNote(ctx context.Context, req UpdateNoteRequest) error

	// DeleteNote deletes a note.
	DeleteNote(ctx context.Context, noteID string) error

	// PinNote pins a note for visibility.
	PinNote(ctx context.Context, noteID string) error

	// UnpinNote unpins a note.
	UnpinNote(ctx context.Context, noteID string) error

	// GetNotesByContainer retrieves notes for a specific container.
	GetNotesByContainer(ctx context.Context, containerType, containerID string) ([]*Note, error)

	// CloseNote closes a note with a reason.
	CloseNote(ctx context.Context, req CloseNoteRequest) error

	// ReopenNote reopens a closed note.
	ReopenNote(ctx context.Context, noteID string) error

	// MoveNote moves a note to a different container.
	MoveNote(ctx context.Context, req MoveNoteRequest) error

	// MergeNotes merges source note into target and closes source.
	MergeNotes(ctx context.Context, req MergeNoteRequest) error
}

// CreateNoteRequest contains parameters for creating a note.
type CreateNoteRequest struct {
	CommissionID  string
	Title         string
	Content       string
	Type          string // learning, concern, finding, frq, bug
	ContainerID   string // The container ID (shipment, conclave, or tome)
	ContainerType string // "shipment", "conclave", or "tome"
}

// CreateNoteResponse contains the result of creating a note.
type CreateNoteResponse struct {
	NoteID string
	Note   *Note
}

// UpdateNoteRequest contains parameters for updating a note.
type UpdateNoteRequest struct {
	NoteID  string
	Title   string
	Content string
	Type    string
}

// MoveNoteRequest contains parameters for moving a note to a different container.
// Exactly one of the To* fields should be set.
type MoveNoteRequest struct {
	NoteID       string
	ToTomeID     string
	ToShipmentID string
	ToConclaveID string
}

// MergeNoteRequest contains parameters for merging one note into another.
type MergeNoteRequest struct {
	SourceNoteID string
	TargetNoteID string
}

// CloseNoteRequest contains parameters for closing a note with a reason.
type CloseNoteRequest struct {
	NoteID   string
	Reason   string // superseded, synthesized, resolved, deferred, duplicate, stale
	ByNoteID string // Optional reference to another note
}

// Note represents a note entity at the port boundary.
type Note struct {
	ID               string
	CommissionID     string
	Title            string
	Content          string
	Type             string
	Status           string // "open" or "closed"
	ShipmentID       string
	ConclaveID       string
	TomeID           string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ClosedAt         string
	PromotedFromID   string
	PromotedFromType string
	CloseReason      string
	ClosedByNoteID   string
}

// NoteFilters contains filter options for listing notes.
type NoteFilters struct {
	Type         string
	CommissionID string
}

// Note type constants
const (
	NoteTypeLearning = "learning"
	NoteTypeConcern  = "concern"
	NoteTypeFinding  = "finding"
	NoteTypeFRQ      = "frq"
	NoteTypeBug      = "bug"
	NoteTypeSpec     = "spec"
	NoteTypeRoadmap  = "roadmap"
	NoteTypeDecision = "decision"
	NoteTypeQuestion = "question"
	NoteTypeVision   = "vision"
	NoteTypeIdea     = "idea"
	NoteTypeExorcism = "exorcism"
)

// Note status constants
const (
	NoteStatusOpen   = "open"
	NoteStatusClosed = "closed"
)
