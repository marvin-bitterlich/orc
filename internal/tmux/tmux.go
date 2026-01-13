package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// Session represents a TMux session
type Session struct {
	Name string
}

// Window represents a TMux window
type Window struct {
	Session *Session
	Index   int
	Name    string
}

// NewSession creates a new TMux session
func NewSession(name, workingDir string) (*Session, error) {
	// Create session with first window, start numbering from 1
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", workingDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set base-index to 1 for this session (windows start at 1)
	exec.Command("tmux", "set-option", "-t", name, "base-index", "1").Run()
	// Set pane-base-index to 1 (panes start at 1)
	exec.Command("tmux", "set-option", "-t", name, "pane-base-index", "1").Run()

	return &Session{Name: name}, nil
}

// KillSession terminates a TMux session
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}

// CreateDeputyWindow creates the deputy control window with claude running
func (s *Session) CreateDeputyWindow() (*Window, error) {
	// First window is already created, just rename it
	target := fmt.Sprintf("%s:1", s.Name)

	if err := exec.Command("tmux", "rename-window", "-t", target, "deputy").Run(); err != nil {
		return nil, fmt.Errorf("failed to rename deputy window: %w", err)
	}

	// Launch claude in the deputy pane
	if err := s.SendKeys(target, "claude"); err != nil {
		return nil, fmt.Errorf("failed to launch claude: %w", err)
	}

	return &Window{Session: s, Index: 1, Name: "deputy"}, nil
}

// CreateMasterOrcWindow creates the master orchestrator window with layout:
// Layout:
//   ┌─────────────────────┬──────────────┐
//   │                     │   vim (top)  │
//   │      claude         │──────────────│
//   │    (full height)    │  shell (bot) │
//   │                     │              │
//   └─────────────────────┴──────────────┘
func (s *Session) CreateMasterOrcWindow(workingDir string) error {
	// First window is already created (window 1), rename it
	target := fmt.Sprintf("%s:1", s.Name)

	if err := exec.Command("tmux", "rename-window", "-t", target, "orc").Run(); err != nil {
		return fmt.Errorf("failed to rename master window: %w", err)
	}

	// Split vertically (creates pane on the right)
	if err := s.SplitVertical(target, workingDir); err != nil {
		return err
	}

	// Now split the right pane horizontally
	// Target the right pane (pane 2)
	rightPane := fmt.Sprintf("%s.2", target)
	if err := s.SplitHorizontal(rightPane, workingDir); err != nil {
		return err
	}

	// Now we have 3 panes:
	// Pane 1 (left): claude (orchestrator)
	// Pane 2 (top right): vim
	// Pane 3 (bottom right): shell

	// Launch claude in pane 1 (left)
	pane1 := fmt.Sprintf("%s.1", target)
	if err := s.SendKeys(pane1, "claude"); err != nil {
		return fmt.Errorf("failed to launch claude: %w", err)
	}

	// Launch vim in pane 2 (top right)
	pane2 := fmt.Sprintf("%s.2", target)
	if err := s.SendKeys(pane2, "vim"); err != nil {
		return fmt.Errorf("failed to launch vim: %w", err)
	}

	// Pane 3 (bottom right) is just a shell, already there

	return nil
}

// CreateGroveWindow creates a grove window with sophisticated layout:
// Layout:
//   ┌─────────────────┬─────────────────┐
//   │                 │ claude (IMP)    │
//   │ vim             ├─────────────────┤
//   │                 │ shell           │
//   └─────────────────┴─────────────────┘
func (s *Session) CreateGroveWindow(index int, name, workingDir string) (*Window, error) {
	// Create new window
	cmd := exec.Command("tmux", "new-window", "-t", s.Name, "-n", name, "-c", workingDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create grove window: %w", err)
	}

	target := fmt.Sprintf("%s:%s", s.Name, name)

	// Get the pane ID for the first pane (will be vim)
	// Split vertically (creates pane on the right)
	if err := s.SplitVertical(target, workingDir); err != nil {
		return nil, err
	}

	// Now split the right pane horizontally
	// Target the right pane (pane 2)
	rightPane := fmt.Sprintf("%s.2", target)
	if err := s.SplitHorizontal(rightPane, workingDir); err != nil {
		return nil, err
	}

	// Now we have 3 panes:
	// Pane 1 (left): vim
	// Pane 2 (top right): claude (IMP)
	// Pane 3 (bottom right): shell

	// Launch vim in pane 1 (left)
	pane1 := fmt.Sprintf("%s.1", target)
	if err := s.SendKeys(pane1, "vim"); err != nil {
		return nil, fmt.Errorf("failed to launch vim: %w", err)
	}

	// Launch claude in pane 2 (top right - IMP)
	pane2 := fmt.Sprintf("%s.2", target)
	if err := s.SendKeys(pane2, "claude"); err != nil {
		return nil, fmt.Errorf("failed to launch claude IMP: %w", err)
	}

	// Pane 3 (bottom right) is just a shell, already there

	return &Window{Session: s, Index: index, Name: name}, nil
}

// SplitVertical splits a pane vertically (creates pane on the right)
func (s *Session) SplitVertical(target, workingDir string) error {
	cmd := exec.Command("tmux", "split-window", "-h", "-t", target, "-c", workingDir)
	return cmd.Run()
}

// SplitHorizontal splits a pane horizontally (creates pane below)
func (s *Session) SplitHorizontal(target, workingDir string) error {
	cmd := exec.Command("tmux", "split-window", "-v", "-t", target, "-c", workingDir)
	return cmd.Run()
}

// SendKeys sends keystrokes to a pane (with Enter)
func (s *Session) SendKeys(target, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys, "C-m")
	return cmd.Run()
}

// SelectWindow switches to a specific window
func (s *Session) SelectWindow(windowIndex int) error {
	target := fmt.Sprintf("%s:%d", s.Name, windowIndex)
	cmd := exec.Command("tmux", "select-window", "-t", target)
	return cmd.Run()
}

// GetSessionInfo returns formatted information about the session
func GetSessionInfo(name string) (string, error) {
	cmd := exec.Command("tmux", "list-windows", "-t", name)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get session info: %w", err)
	}
	return string(output), nil
}

// SessionExists checks if a TMux session exists
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	err := cmd.Run()
	return err == nil
}

// AttachInstructions returns user-friendly instructions for attaching to session
func AttachInstructions(sessionName string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Attach to session: tmux attach -t %s\n", sessionName))
	b.WriteString("\n")
	b.WriteString("Window Layout:\n")
	b.WriteString("  Window 1 (deputy): Claude for mission coordination\n")
	b.WriteString("  Windows 2+: Grove workspaces with vim, claude IMP, and shell\n")
	b.WriteString("\n")
	b.WriteString("TMux Commands:\n")
	b.WriteString("  Switch windows: Ctrl+b then window number (1, 2, 3...)\n")
	b.WriteString("  Switch panes: Ctrl+b then arrow keys\n")
	b.WriteString("  Detach session: Ctrl+b then d\n")
	b.WriteString("  List windows: Ctrl+b then w\n")

	return b.String()
}
