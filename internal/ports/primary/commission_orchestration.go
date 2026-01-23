package primary

// CommissionState represents the loaded state of a commission.
type CommissionState struct {
	Commission *Commission
}

// InfrastructurePlan describes the changes needed to set up commission infrastructure.
type InfrastructurePlan struct {
	WorkspacePath  string
	WorkbenchesDir string

	// Actions to perform
	CreateWorkspace      bool
	CreateWorkbenchesDir bool
	ConfigWrites         []ConfigWrite
	Cleanups             []CleanupAction
}

// ConfigWrite represents a config file to write.
type ConfigWrite struct {
	Path    string
	Type    string // "workbench", "claude-settings"
	Content string // For preview only
}

// CleanupAction represents a file to clean up.
type CleanupAction struct {
	Path   string
	Reason string
}

// InfrastructureApplyResult captures the result of applying infrastructure changes.
type InfrastructureApplyResult struct {
	WorkspaceCreated      bool
	WorkbenchesDirCreated bool
	WorkbenchesProcessed  int
	ConfigsWritten        int
	CleanupsDone          int
	Errors                []string
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
	Index         int
	Name          string
	WorkbenchID   string
	WorkbenchPath string
	Action        string // "create", "exists", "update", "skip"
	PaneCount     int
	NeedsUpdate   bool
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
