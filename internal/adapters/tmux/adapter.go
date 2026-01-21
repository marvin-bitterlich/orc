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

// CreateGroveWindow creates a grove window with IMP workspace layout.
func (a *Adapter) CreateGroveWindow(ctx context.Context, sessionName string, windowIndex int, windowName, workingDir string) error {
	session := &tmuxpkg.Session{Name: sessionName}
	_, err := session.CreateGroveWindow(windowIndex, windowName, workingDir)
	return err
}

// CreateGroveWindowShell creates a grove window with shell layout (no apps launched).
func (a *Adapter) CreateGroveWindowShell(ctx context.Context, sessionName string, windowIndex int, windowName, workingDir string) error {
	session := &tmuxpkg.Session{Name: sessionName}
	_, err := session.CreateGroveWindowShell(windowIndex, windowName, workingDir)
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

// Ensure Adapter implements the interface
var _ secondary.TMuxAdapter = (*Adapter)(nil)
