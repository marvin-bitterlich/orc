// Package tmux contains TMux adapter implementations.
package tmux

import (
	"context"

	"github.com/example/orc/internal/ports/secondary"
	tmuxpkg "github.com/example/orc/internal/tmux"
)

// Adapter implements secondary.TMuxAdapter by wrapping the internal/tmux package.
type Adapter struct{}

// NewAdapter creates a new TMux adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// SessionExists checks if a TMux session exists.
func (a *Adapter) SessionExists(ctx context.Context, name string) bool {
	return tmuxpkg.SessionExists(name)
}

// KillSession terminates a TMux session.
func (a *Adapter) KillSession(ctx context.Context, name string) error {
	return tmuxpkg.KillSession(name)
}

// GetSessionInfo returns information about a TMux session.
func (a *Adapter) GetSessionInfo(ctx context.Context, name string) (string, error) {
	return tmuxpkg.GetSessionInfo(name)
}

// WindowExists checks if a window exists in a session.
func (a *Adapter) WindowExists(ctx context.Context, sessionName, windowName string) bool {
	return tmuxpkg.WindowExists(sessionName, windowName)
}

// KillWindow kills a window in a session.
func (a *Adapter) KillWindow(ctx context.Context, sessionName, windowName string) error {
	return tmuxpkg.KillWindow(sessionName, windowName)
}

// SendKeys sends keystrokes to a pane.
func (a *Adapter) SendKeys(ctx context.Context, target, keys string) error {
	session := &tmuxpkg.Session{Name: ""} // Name not needed for SendKeys
	return session.SendKeys(target, keys)
}

// GetPaneCount returns the number of panes in a window.
func (a *Adapter) GetPaneCount(ctx context.Context, sessionName, windowName string) int {
	return tmuxpkg.GetPaneCount(sessionName, windowName)
}

// GetPaneCommand returns the current command running in a pane.
func (a *Adapter) GetPaneCommand(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return tmuxpkg.GetPaneCommand(sessionName, windowName, paneNum)
}

// GetPaneStartPath returns the initial directory a pane was created with.
func (a *Adapter) GetPaneStartPath(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return tmuxpkg.GetPaneStartPath(sessionName, windowName, paneNum)
}

// GetPaneStartCommand returns the initial command a pane was created with (via respawn-pane).
func (a *Adapter) GetPaneStartCommand(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return tmuxpkg.GetPaneStartCommand(sessionName, windowName, paneNum)
}

// CapturePaneContent captures visible content from a pane.
func (a *Adapter) CapturePaneContent(ctx context.Context, target string, lines int) (string, error) {
	return tmuxpkg.CapturePaneContent(target, lines)
}

// AttachInstructions returns user-friendly instructions for attaching to a session.
func (a *Adapter) AttachInstructions(sessionName string) string {
	return tmuxpkg.AttachInstructions(sessionName)
}

// SplitVertical splits a pane vertically.
func (a *Adapter) SplitVertical(ctx context.Context, target, workingDir string) error {
	session := &tmuxpkg.Session{Name: ""}
	return session.SplitVertical(target, workingDir)
}

// SplitHorizontal splits a pane horizontally.
func (a *Adapter) SplitHorizontal(ctx context.Context, target, workingDir string) error {
	session := &tmuxpkg.Session{Name: ""}
	return session.SplitHorizontal(target, workingDir)
}

// JoinPane moves a pane from source to target.
func (a *Adapter) JoinPane(ctx context.Context, source, target string, vertical bool, size int) error {
	return tmuxpkg.JoinPane(source, target, vertical, size)
}

// SelectWindow selects a window by index.
func (a *Adapter) SelectWindow(ctx context.Context, sessionName string, index int) error {
	session := &tmuxpkg.Session{Name: sessionName}
	return session.SelectWindow(index)
}

// RenameWindow renames a window.
func (a *Adapter) RenameWindow(ctx context.Context, target, newName string) error {
	return tmuxpkg.RenameWindow(target, newName)
}

// RespawnPane respawns a pane with optional command.
func (a *Adapter) RespawnPane(ctx context.Context, target string, command ...string) error {
	return tmuxpkg.RespawnPane(target, command...)
}

// RenameSession renames a TMux session.
func (a *Adapter) RenameSession(ctx context.Context, session, newName string) error {
	return tmuxpkg.RenameSession(session, newName)
}

// ConfigureStatusBar configures the TMux status bar.
func (a *Adapter) ConfigureStatusBar(ctx context.Context, session string, config secondary.StatusBarConfig) error {
	if config.StatusLeft != "" {
		if err := tmuxpkg.SetOption(session, "status-left", config.StatusLeft); err != nil {
			return err
		}
	}
	if config.StatusRight != "" {
		if err := tmuxpkg.SetOption(session, "status-right", config.StatusRight); err != nil {
			return err
		}
	}
	return nil
}

// DisplayPopup displays a popup in a TMux session.
func (a *Adapter) DisplayPopup(ctx context.Context, session, command string, config secondary.PopupConfig) error {
	return tmuxpkg.DisplayPopup(session, command, config.Width, config.Height, config.Title)
}

// ConfigureSessionBindings sets up key bindings for a session.
func (a *Adapter) ConfigureSessionBindings(ctx context.Context, session string, bindings []secondary.KeyBinding) error {
	for _, b := range bindings {
		if err := tmuxpkg.BindKey(session, b.Key, b.Command); err != nil {
			return err
		}
	}
	return nil
}

// ConfigureSessionPopupBindings sets up key bindings that display popups.
func (a *Adapter) ConfigureSessionPopupBindings(ctx context.Context, session string, bindings []secondary.PopupKeyBinding) error {
	for _, b := range bindings {
		if err := tmuxpkg.BindKeyPopup(session, b.Key, b.Command, b.Config.Width, b.Config.Height, b.Config.Title, b.Config.WorkingDir); err != nil {
			return err
		}
	}
	return nil
}

// GetCurrentSessionName returns the name of the current tmux session.
func (a *Adapter) GetCurrentSessionName(ctx context.Context) string {
	return tmuxpkg.GetCurrentSessionName()
}

// SetEnvironment sets an environment variable for a tmux session.
func (a *Adapter) SetEnvironment(ctx context.Context, sessionName, key, value string) error {
	return tmuxpkg.SetEnvironment(sessionName, key, value)
}

// GetEnvironment gets an environment variable from a tmux session.
func (a *Adapter) GetEnvironment(ctx context.Context, sessionName, key string) (string, error) {
	return tmuxpkg.GetEnvironment(sessionName, key)
}

// ListSessions returns all tmux session names.
func (a *Adapter) ListSessions(ctx context.Context) ([]string, error) {
	return tmuxpkg.ListSessions()
}

// FindSessionByWorkshopID finds the session with ORC_WORKSHOP_ID=workshopID.
func (a *Adapter) FindSessionByWorkshopID(ctx context.Context, workshopID string) string {
	return tmuxpkg.FindSessionByWorkshopID(workshopID)
}

// ListWindows returns window names in a session.
func (a *Adapter) ListWindows(ctx context.Context, sessionName string) ([]string, error) {
	return tmuxpkg.ListWindows(sessionName)
}

// GetWindowOption gets a window option value.
func (a *Adapter) GetWindowOption(ctx context.Context, target, option string) string {
	return tmuxpkg.GetWindowOption(target, option)
}

// SetWindowOption sets a window option value.
func (a *Adapter) SetWindowOption(ctx context.Context, target, option, value string) error {
	return tmuxpkg.SetWindowOption(target, option, value)
}

// SetupGoblinPane launches orc connect --role goblin in pane 1 of an existing window.
func (a *Adapter) SetupGoblinPane(ctx context.Context, sessionName, windowName string) error {
	target := sessionName + ":" + windowName
	return tmuxpkg.SetupGoblinPane(target)
}

// ApplyGlobalBindings sets up ORC's global tmux key bindings.
// Safe to call repeatedly (idempotent). Silently ignores errors (tmux may not be running).
func ApplyGlobalBindings() {
	tmuxpkg.ApplyGlobalBindings()
}

// RefreshWorkbenchLayout relocates guest panes (no PANE_ROLE) to a sibling -imps window.
// Non-destructive - guest processes keep running, just moved to a separate window.
func RefreshWorkbenchLayout(sessionName, workbenchWindow string) error {
	return tmuxpkg.RefreshWorkbenchLayout(sessionName, workbenchWindow)
}

// EnrichSession applies ORC enrichment to all windows in a session.
// This includes setting PANE_ROLE env vars, pane titles, and window options.
func EnrichSession(sessionName string) error {
	return tmuxpkg.EnrichSession(sessionName)
}

// GotmuxAdapter re-exports gotmux adapter for wire injection.
type GotmuxAdapter = tmuxpkg.GotmuxAdapter

// DesiredWorkbench re-exports the type for plan building.
type DesiredWorkbench = tmuxpkg.DesiredWorkbench

// ApplyPlan re-exports the reconciliation plan type.
type ApplyPlan = tmuxpkg.ApplyPlan

// NewGotmuxAdapter creates a new gotmux adapter.
func NewGotmuxAdapter() (*GotmuxAdapter, error) {
	return tmuxpkg.NewGotmuxAdapter()
}

// Ensure Adapter implements the interface
var _ secondary.TMuxAdapter = (*Adapter)(nil)
