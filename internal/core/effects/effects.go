// Package effects defines effect types as data structures representing I/O operations.
// This is the foundation of the Functional Core / Imperative Shell pattern.
// Effects are pure data - they describe what should happen, not how.
package effects

// Effect is the base interface for all effects.
// Effects represent I/O operations as data that can be interpreted by the shell.
type Effect interface {
	// EffectType returns a string identifier for the effect type.
	EffectType() string
}

// LogEffect represents a logging operation.
type LogEffect struct {
	Level   string
	Message string
	Fields  map[string]any
}

func (e LogEffect) EffectType() string { return "log" }

// PersistEffect represents a database persistence operation.
type PersistEffect struct {
	Entity     string // e.g., "mission", "grove", "work_order"
	Operation  string // e.g., "create", "update", "delete"
	Data       any    // The entity data
	Conditions any    // Optional conditions/filters
}

func (e PersistEffect) EffectType() string { return "persist" }

// QueryEffect represents a database query operation.
type QueryEffect struct {
	Entity     string
	Conditions any
	OrderBy    string
	Limit      int
}

func (e QueryEffect) EffectType() string { return "query" }

// FileEffect represents a file system operation.
type FileEffect struct {
	Operation string // e.g., "read", "write", "mkdir", "exists"
	Path      string
	Content   []byte // For write operations
	Mode      uint32 // File permissions
}

func (e FileEffect) EffectType() string { return "file" }

// GitEffect represents a git operation.
type GitEffect struct {
	Operation string   // e.g., "clone", "worktree_add", "commit", "push"
	RepoPath  string   // Path to repository
	Args      []string // Additional arguments
}

func (e GitEffect) EffectType() string { return "git" }

// TMuxEffect represents a tmux operation.
type TMuxEffect struct {
	Operation   string // e.g., "new_session", "new_window", "send_keys"
	SessionName string
	WindowName  string
	Command     string
}

func (e TMuxEffect) EffectType() string { return "tmux" }

// CompositeEffect holds multiple effects to be executed in sequence.
type CompositeEffect struct {
	Effects []Effect
}

func (e CompositeEffect) EffectType() string { return "composite" }

// NoEffect represents an operation that produces no side effects.
type NoEffect struct{}

func (e NoEffect) EffectType() string { return "none" }
