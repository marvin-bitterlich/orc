package primary

// MissionState represents the loaded state of a mission and its groves.
type MissionState struct {
	Mission *Mission
	Groves  []*Grove
}

// InfrastructurePlan describes the changes needed to set up mission infrastructure.
type InfrastructurePlan struct {
	WorkspacePath string
	GrovesDir     string

	// Actions to perform
	CreateWorkspace bool
	CreateGrovesDir bool
	GroveActions    []GroveAction
	ConfigWrites    []ConfigWrite
	Cleanups        []CleanupAction
}

// GroveAction represents an action to take on a grove.
type GroveAction struct {
	GroveID      string
	GroveName    string
	CurrentPath  string
	DesiredPath  string
	Action       string // "exists", "create", "move", "missing"
	PathExists   bool
	UpdateDBPath bool
}

// ConfigWrite represents a config file to write.
type ConfigWrite struct {
	Path    string
	Type    string // "grove", "claude-settings"
	Grove   *Grove
	Content string // For preview only
}

// CleanupAction represents a file to clean up.
type CleanupAction struct {
	Path   string
	Reason string
}

// InfrastructureApplyResult captures the result of applying infrastructure changes.
type InfrastructureApplyResult struct {
	WorkspaceCreated  bool
	GrovesDirCreated  bool
	GrovesProcessed   int
	ConfigsWritten    int
	CleanupsDone      int
	Errors            []string
	GrovesNeedingWork []GroveNeedingWork
}

// GroveNeedingWork describes a grove that needs additional work.
type GroveNeedingWork struct {
	GroveID     string
	GroveName   string
	DesiredPath string
	Message     string
}

// TmuxSessionPlan describes the TMux session to create or update.
type TmuxSessionPlan struct {
	SessionName   string
	WorkingDir    string
	SessionExists bool
	WindowPlans   []WindowPlan
}

// WindowPlan describes a TMux window to create or update.
type WindowPlan struct {
	Index       int
	Name        string
	GroveID     string
	GrovePath   string
	Action      string // "create", "exists", "update", "skip"
	PaneCount   int
	NeedsUpdate bool
}

// TmuxSessionResult captures the result of applying TMux session changes.
type TmuxSessionResult struct {
	SessionCreated bool
	WindowsCreated int
	WindowsUpdated int
	Errors         []string
}

// TmuxWindowChecker is an interface for checking TMux window state.
// This allows dependency injection for testing.
type TmuxWindowChecker interface {
	WindowExists(session, window string) bool
	GetPaneCount(session, window string) int
	GetPaneCommand(session, window string, pane int) string
}
