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

// CreateSession creates a new TMux session.
func (a *Adapter) CreateSession(ctx context.Context, name, workingDir string) error {
	_, err := tmuxpkg.NewSession(name, workingDir)
	return err
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

// CreateOrcWindow creates the ORC orchestrator window layout.
func (a *Adapter) CreateOrcWindow(ctx context.Context, sessionName, workingDir string) error {
	session := &tmuxpkg.Session{Name: sessionName}
	return session.CreateOrcWindow(workingDir)
}

// CreateWorkbenchWindow creates a workbench window with IMP workspace layout.
func (a *Adapter) CreateWorkbenchWindow(ctx context.Context, sessionName string, windowIndex int, windowName, workingDir string) error {
	session := &tmuxpkg.Session{Name: sessionName}
	_, err := session.CreateWorkbenchWindow(windowIndex, windowName, workingDir)
	return err
}

// CreateWorkbenchWindowShell creates a workbench window with shell layout (no apps launched).
func (a *Adapter) CreateWorkbenchWindowShell(ctx context.Context, sessionName string, windowIndex int, windowName, workingDir string) error {
	session := &tmuxpkg.Session{Name: sessionName}
	_, err := session.CreateWorkbenchWindowShell(windowIndex, windowName, workingDir)
	return err
}

// WindowExists checks if a window exists in a session.
func (a *Adapter) WindowExists(ctx context.Context, sessionName, windowName string) bool {
	return tmuxpkg.WindowExists(sessionName, windowName)
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

// NudgeSession sends a message to a running Claude session.
func (a *Adapter) NudgeSession(ctx context.Context, target, message string) error {
	return tmuxpkg.NudgeSession(target, message)
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

// Ensure Adapter implements the interface
var _ secondary.TMuxAdapter = (*Adapter)(nil)
