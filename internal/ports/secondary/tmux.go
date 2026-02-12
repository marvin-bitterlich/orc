// Package secondary defines the secondary ports (driven adapters) for the application.
package secondary

import "context"

// StatusBarConfig configures tmux status bar appearance.
type StatusBarConfig struct {
	StatusLeft  string // format string for left side
	StatusRight string // format string for right side
}

// PopupConfig configures tmux popup display.
type PopupConfig struct {
	Width      int    // popup width (columns)
	Height     int    // popup height (rows)
	Title      string // optional popup title
	WorkingDir string // working directory (supports tmux format strings like #{pane_current_path})
}

// KeyBinding defines a tmux key binding.
type KeyBinding struct {
	Key     string // e.g., "MouseDown3Status"
	Command string // tmux command to execute
}

// PopupKeyBinding defines a key binding that shows a popup.
type PopupKeyBinding struct {
	Key     string      // e.g., "DoubleClick1Status"
	Command string      // command to run in popup
	Config  PopupConfig // popup dimensions/title
}

// TMuxAdapter defines the secondary port for TMux session and window management.
type TMuxAdapter interface {
	// Session management
	SessionExists(ctx context.Context, name string) bool
	KillSession(ctx context.Context, name string) error
	GetSessionInfo(ctx context.Context, name string) (string, error)

	// Window management
	WindowExists(ctx context.Context, sessionName string, windowName string) bool
	KillWindow(ctx context.Context, sessionName string, windowName string) error

	// Pane operations
	SendKeys(ctx context.Context, target, keys string) error
	GetPaneCount(ctx context.Context, sessionName, windowName string) int
	GetPaneCommand(ctx context.Context, sessionName, windowName string, paneNum int) string
	GetPaneStartPath(ctx context.Context, sessionName, windowName string, paneNum int) string
	GetPaneStartCommand(ctx context.Context, sessionName, windowName string, paneNum int) string
	CapturePaneContent(ctx context.Context, target string, lines int) (string, error)
	SplitVertical(ctx context.Context, target, workingDir string) error
	SplitHorizontal(ctx context.Context, target, workingDir string) error
	JoinPane(ctx context.Context, source, target string, vertical bool, size int) error

	// Information
	AttachInstructions(sessionName string) string

	// Window navigation
	SelectWindow(ctx context.Context, sessionName string, index int) error
	RenameWindow(ctx context.Context, target, newName string) error
	RespawnPane(ctx context.Context, target string, command ...string) error

	// UI operations
	RenameSession(ctx context.Context, session, newName string) error
	ConfigureStatusBar(ctx context.Context, session string, config StatusBarConfig) error
	DisplayPopup(ctx context.Context, session, command string, config PopupConfig) error
	ConfigureSessionBindings(ctx context.Context, session string, bindings []KeyBinding) error
	ConfigureSessionPopupBindings(ctx context.Context, session string, bindings []PopupKeyBinding) error

	// Session info
	GetCurrentSessionName(ctx context.Context) string

	// Environment variables
	SetEnvironment(ctx context.Context, sessionName, key, value string) error
	GetEnvironment(ctx context.Context, sessionName, key string) (string, error)

	// Session discovery
	ListSessions(ctx context.Context) ([]string, error)
	FindSessionByWorkshopID(ctx context.Context, workshopID string) string
	ListWindows(ctx context.Context, sessionName string) ([]string, error)

	// Window options (for @orc_agent, @orc_focus, etc.)
	GetWindowOption(ctx context.Context, target, option string) string
	SetWindowOption(ctx context.Context, target, option, value string) error

	// Goblin window setup
	SetupGoblinPane(ctx context.Context, sessionName, windowName string) error
}
