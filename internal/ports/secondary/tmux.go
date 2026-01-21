// Package secondary defines the secondary ports (driven adapters) for the application.
package secondary

import "context"

// TMuxAdapter defines the secondary port for TMux session and window management.
type TMuxAdapter interface {
	// Session management
	CreateSession(ctx context.Context, name, workingDir string) error
	SessionExists(ctx context.Context, name string) bool
	KillSession(ctx context.Context, name string) error
	GetSessionInfo(ctx context.Context, name string) (string, error)

	// Window management
	CreateOrcWindow(ctx context.Context, sessionName string, workingDir string) error
	CreateGroveWindow(ctx context.Context, sessionName string, windowIndex int, windowName string, workingDir string) error
	CreateGroveWindowShell(ctx context.Context, sessionName string, windowIndex int, windowName string, workingDir string) error
	WindowExists(ctx context.Context, sessionName string, windowName string) bool

	// Pane operations
	SendKeys(ctx context.Context, target, keys string) error
	GetPaneCount(ctx context.Context, sessionName, windowName string) int
	GetPaneCommand(ctx context.Context, sessionName, windowName string, paneNum int) string

	// Communication
	NudgeSession(ctx context.Context, target, message string) error

	// Information
	AttachInstructions(sessionName string) string
}
