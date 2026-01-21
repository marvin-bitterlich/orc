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

// WindowExists checks if a window exists in a session
func WindowExists(sessionName, windowName string) bool {
	cmd := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_name}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	windows := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, w := range windows {
		if w == windowName {
			return true
		}
	}
	return false
}

// GetPaneCount returns the number of panes in a window
func GetPaneCount(sessionName, windowName string) int {
	target := fmt.Sprintf("%s:%s", sessionName, windowName)
	cmd := exec.Command("tmux", "list-panes", "-t", target)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return len(lines)
}

// GetPaneCommand returns the current command running in a specific pane
// Returns empty string if pane doesn't exist or error occurs
func GetPaneCommand(sessionName, windowName string, paneNum int) string {
	target := fmt.Sprintf("%s:%s.%d", sessionName, windowName, paneNum)
	cmd := exec.Command("tmux", "display-message", "-t", target, "-p", "#{pane_current_command}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// CreateOrcWindow creates the ORC orchestrator window with layout:
// Layout:
//
//	┌─────────────────────┬──────────────┐
//	│                     │   vim (top)  │
//	│      claude         │──────────────│
//	│    (full height)    │  shell (bot) │
//	│                     │              │
//	└─────────────────────┴──────────────┘
func (s *Session) CreateOrcWindow(workingDir string) error {
	// First window is already created (window 1), rename it
	target := fmt.Sprintf("%s:1", s.Name)

	if err := exec.Command("tmux", "rename-window", "-t", target, "orc").Run(); err != nil {
		return fmt.Errorf("failed to rename ORC window: %w", err)
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

	// Launch claude in pane 1 (left) with orc prime prompt
	pane1 := fmt.Sprintf("%s.1", target)
	if err := s.SendKeys(pane1, "claude \"Run the orc prime command to get context\""); err != nil {
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

// CreateGroveWindowShell creates a grove window with layout but NO app launching
// Layout:
//
//	┌─────────────────┬─────────────────┐
//	│                 │ (top right)     │
//	│ (left pane)     ├─────────────────┤
//	│                 │ (bottom right)  │
//	└─────────────────┴─────────────────┘
//
// Apps (vim, claude) can be launched later
func (s *Session) CreateGroveWindowShell(index int, name, workingDir string) (*Window, error) {
	// Create new window
	cmd := exec.Command("tmux", "new-window", "-t", s.Name, "-n", name, "-c", workingDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create grove window: %w", err)
	}

	target := fmt.Sprintf("%s:%s", s.Name, name)

	// Split vertically (creates pane on the right)
	if err := s.SplitVertical(target, workingDir); err != nil {
		return nil, err
	}

	// Split the right pane horizontally
	rightPane := fmt.Sprintf("%s.2", target)
	if err := s.SplitHorizontal(rightPane, workingDir); err != nil {
		return nil, err
	}

	// Now we have 3 panes ready:
	// Pane 1 (left): shell (for vim)
	// Pane 2 (top right): will become IMP (orc connect)
	// Pane 3 (bottom right): shell

	// Launch orc connect in top-right pane (pane 2)
	// Using respawn-pane makes "orc connect" the root command
	// This means if the pane exits or is respawned, it runs orc connect again
	topRightPane := fmt.Sprintf("%s.2", target)
	connectCmd := exec.Command("tmux", "respawn-pane", "-t", topRightPane, "-k", "orc", "connect")
	if err := connectCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to launch orc connect in top-right pane: %w", err)
	}

	return &Window{Session: s, Index: index, Name: name}, nil
}

// CreateGroveWindow creates a grove window with sophisticated layout:
// Layout:
//
//	┌─────────────────┬─────────────────┐
//	│                 │ claude (IMP)    │
//	│ vim             ├─────────────────┤
//	│                 │ shell           │
//	└─────────────────┴─────────────────┘
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

	// Launch claude in pane 2 (top right - IMP) with orc prime prompt
	pane2 := fmt.Sprintf("%s.2", target)
	if err := s.SendKeys(pane2, "claude \"Run the orc prime command to get context\""); err != nil {
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
	b.WriteString("  Window 1 (orc): ORC orchestrator (claude | vim | shell)\n")
	b.WriteString("  Windows 2+: Grove workspaces (vim | claude IMP | shell)\n")
	b.WriteString("\n")
	b.WriteString("TMux Commands:\n")
	b.WriteString("  Switch windows: Ctrl+b then window number (1, 2, 3...)\n")
	b.WriteString("  Switch panes: Ctrl+b then arrow keys\n")
	b.WriteString("  Detach session: Ctrl+b then d\n")
	b.WriteString("  List windows: Ctrl+b then w\n")

	return b.String()
}

// SendKeysLiteral sends text literally without interpretation
func (s *Session) SendKeysLiteral(target, text string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, "-l", text)
	return cmd.Run()
}

// SendEscape sends the Escape key
func (s *Session) SendEscape(target string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, "Escape")
	return cmd.Run()
}

// SendEnter sends the Enter key
func (s *Session) SendEnter(target string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, "Enter")
	return cmd.Run()
}

// NudgeSession sends a message to a running Claude session using the Gastown pattern
// This implements the 4-step reliable delivery:
// 1. Send text literally (no interpretation)
// 2. Wait 500ms for processing
// 3. Send Escape to exit vim mode
// 4. Send Enter to submit (with retry logic)
func NudgeSession(target, message string) error {
	// Extract session name from target
	parts := strings.Split(target, ":")
	if len(parts) == 0 {
		return fmt.Errorf("invalid target format: %s", target)
	}
	session := &Session{Name: parts[0]}

	// Step 1: Send text literally
	if err := session.SendKeysLiteral(target, message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Step 2: Wait 500ms (critical for reliability)
	cmd := exec.Command("sleep", "0.5")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to wait: %w", err)
	}

	// Step 3: Send Escape for vim mode
	if err := session.SendEscape(target); err != nil {
		return fmt.Errorf("failed to send escape: %w", err)
	}

	// Extra 100ms wait after Escape (from Gastown fix)
	cmd = exec.Command("sleep", "0.1")
	_ = cmd.Run()

	// Step 4: Send Enter to submit with retry (critical fix from Gastown issue #307)
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			cmd := exec.Command("sleep", "0.2")
			_ = cmd.Run()
		}
		if err := session.SendEnter(target); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("failed to send Enter after 3 attempts: %w", lastErr)
}
