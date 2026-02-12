package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// exactSession returns a tmux target string that matches the session name exactly.
// Prefixing with "=" prevents tmux from partial-matching against other sessions.
// Use this for all session-targeted commands to avoid cross-session interference.
func exactSession(sessionName string) string {
	return "=" + sessionName
}

// exactTarget returns a tmux target string for session:window with exact session matching.
func exactTarget(sessionName, windowName string) string {
	return exactSession(sessionName) + ":" + windowName
}

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

	// Rename the auto-created first window to a placeholder
	// The apply logic will rename it to the proper name (e.g., "goblin")
	exec.Command("tmux", "rename-window", "-t", name+":^", "__init__").Run()

	return &Session{Name: name}, nil
}

// KillSession terminates a TMux session
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", exactSession(name))
	return cmd.Run()
}

// WindowExists checks if a window exists in a session
func WindowExists(sessionName, windowName string) bool {
	cmd := exec.Command("tmux", "list-windows", "-t", exactSession(sessionName), "-F", "#{window_name}")
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

// KillWindow kills a window in a session
func KillWindow(sessionName, windowName string) error {
	cmd := exec.Command("tmux", "kill-window", "-t", exactTarget(sessionName, windowName))
	return cmd.Run()
}

// GetPaneCount returns the number of panes in a window
func GetPaneCount(sessionName, windowName string) int {
	target := exactTarget(sessionName, windowName)
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
	target := fmt.Sprintf("=%s:%s.%d", sessionName, windowName, paneNum)
	cmd := exec.Command("tmux", "display-message", "-t", target, "-p", "#{pane_current_command}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetPaneStartPath returns the initial directory for a pane (pane_start_path).
// This is set when the pane is created and does not change.
// Returns empty string if pane doesn't exist or error occurs.
func GetPaneStartPath(sessionName, windowName string, paneNum int) string {
	target := fmt.Sprintf("=%s:%s.%d", sessionName, windowName, paneNum)
	cmd := exec.Command("tmux", "display-message", "-t", target, "-p", "#{pane_start_path}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetPaneStartCommand returns the initial command for a pane (pane_start_command).
// This is only set when the pane is created with respawn-pane or similar.
// Returns empty string if not set, pane doesn't exist, or error occurs.
func GetPaneStartCommand(sessionName, windowName string, paneNum int) string {
	target := fmt.Sprintf("=%s:%s.%d", sessionName, windowName, paneNum)
	cmd := exec.Command("tmux", "display-message", "-t", target, "-p", "#{pane_start_command}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// CapturePaneContent captures visible content from a pane.
// target is in format "session:window.pane" (e.g., "workshop:bench.2")
// lines specifies how many lines to capture (0 for all visible)
func CapturePaneContent(target string, lines int) (string, error) {
	args := []string{"capture-pane", "-t", target, "-p"}
	if lines > 0 {
		args = append(args, "-S", fmt.Sprintf("-%d", lines))
	}
	cmd := exec.Command("tmux", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane content: %w", err)
	}
	return string(output), nil
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

	if err := exec.Command("tmux", "rename-window", "-t", target, "goblin").Run(); err != nil {
		return fmt.Errorf("failed to rename goblin window: %w", err)
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
	// Pane 1 (left): claude (orchestrator via orc connect --role goblin)
	// Pane 2 (top right): vim
	// Pane 3 (bottom right): shell

	// Launch orc connect --role goblin in pane 1 (left) - uses respawn-pane so it's the root command
	pane1 := fmt.Sprintf("%s.1", target)
	connectCmd := exec.Command("tmux", "respawn-pane", "-t", pane1, "-k", "orc", "connect", "--role", "goblin")
	if err := connectCmd.Run(); err != nil {
		return fmt.Errorf("failed to launch orc connect: %w", err)
	}

	// Launch vim in pane 2 (top right)
	pane2 := fmt.Sprintf("%s.2", target)
	if err := s.SendKeys(pane2, "vim"); err != nil {
		return fmt.Errorf("failed to launch vim: %w", err)
	}

	// Pane 3 (bottom right) is just a shell, already there

	return nil
}

// CreateWorkbenchWindowShell creates a workbench window with layout but NO app launching
// Layout:
//
//	┌─────────────────┬─────────────────┐
//	│                 │ (top right)     │
//	│ (left pane)     ├─────────────────┤
//	│                 │ (bottom right)  │
//	└─────────────────┴─────────────────┘
//
// Apps (vim, claude) can be launched later
func (s *Session) CreateWorkbenchWindowShell(index int, name, workingDir string) (*Window, error) {
	// Create new window
	cmd := exec.Command("tmux", "new-window", "-t", s.Name, "-n", name, "-c", workingDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create workbench window: %w", err)
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

// CreateWorkbenchWindow creates a workbench window with sophisticated layout:
// Layout:
//
//	┌─────────────────┬─────────────────┐
//	│                 │ claude (IMP)    │
//	│ vim             ├─────────────────┤
//	│                 │ shell           │
//	└─────────────────┴─────────────────┘
func (s *Session) CreateWorkbenchWindow(index int, name, workingDir string) (*Window, error) {
	// Create new window
	cmd := exec.Command("tmux", "new-window", "-t", s.Name, "-n", name, "-c", workingDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create workbench window: %w", err)
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
	// Pane 2 (top right): claude (IMP via orc connect)
	// Pane 3 (bottom right): shell

	// Launch vim in pane 1 (left) - use respawn-pane so pane_start_command is set
	pane1 := fmt.Sprintf("%s.1", target)
	vimCmd := exec.Command("tmux", "respawn-pane", "-t", pane1, "-k", "vim")
	if err := vimCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to launch vim: %w", err)
	}

	// Launch orc connect in pane 2 (top right - IMP) - uses respawn-pane so it's the root command
	pane2 := fmt.Sprintf("%s.2", target)
	connectCmd := exec.Command("tmux", "respawn-pane", "-t", pane2, "-k", "orc", "connect")
	if err := connectCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to launch orc connect: %w", err)
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

// JoinPane moves a pane from source to target.
// If vertical is true, joins vertically (-v); otherwise horizontally (-h).
// Size specifies the target pane size in lines (if vertical) or columns (if horizontal).
func JoinPane(source, target string, vertical bool, size int) error {
	args := []string{"join-pane"}
	if vertical {
		args = append(args, "-v")
	} else {
		args = append(args, "-h")
	}
	if size > 0 {
		args = append(args, "-l", strconv.Itoa(size))
	}
	args = append(args, "-s", source, "-t", target)
	cmd := exec.Command("tmux", args...)
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

// RenameWindow renames a window
func RenameWindow(target, newName string) error {
	cmd := exec.Command("tmux", "rename-window", "-t", target, newName)
	return cmd.Run()
}

// RespawnPane respawns a pane with optional command
func RespawnPane(target string, command ...string) error {
	args := []string{"respawn-pane", "-t", target, "-k"}
	args = append(args, command...)
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// SetupGoblinPane launches orc connect --role goblin in pane 1 of an existing window.
// Target format: "session:window" (e.g., "WORK-005:goblin")
func SetupGoblinPane(target string) error {
	pane1 := fmt.Sprintf("%s.1", target)
	cmd := exec.Command("tmux", "respawn-pane", "-t", pane1, "-k", "orc", "connect", "--role", "goblin")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch orc connect in goblin pane: %w", err)
	}
	return nil
}

// GetSessionInfo returns formatted information about the session
func GetSessionInfo(name string) (string, error) {
	cmd := exec.Command("tmux", "list-windows", "-t", exactSession(name))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get session info: %w", err)
	}
	return string(output), nil
}

// SessionExists checks if a TMux session exists
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", exactSession(name))
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
	b.WriteString("  Windows 2+: Workbench workspaces (vim | claude IMP | shell)\n")
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

// RenameSession renames a tmux session.
func RenameSession(oldName, newName string) error {
	cmd := exec.Command("tmux", "rename-session", "-t", exactSession(oldName), newName)
	return cmd.Run()
}

// GetCurrentSessionName returns the name of the current tmux session.
// Returns empty string if not in tmux or on error.
func GetCurrentSessionName() string {
	cmd := exec.Command("tmux", "display-message", "-p", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// SetOption sets a tmux option for a session.
func SetOption(session, option, value string) error {
	cmd := exec.Command("tmux", "set-option", "-t", session, option, value)
	return cmd.Run()
}

// DisplayPopup shows a popup window with a command.
func DisplayPopup(session, command string, width, height int, title string) error {
	args := []string{"display-popup", "-t", session, "-E"}
	if width > 0 {
		args = append(args, "-w", strconv.Itoa(width))
	}
	if height > 0 {
		args = append(args, "-h", strconv.Itoa(height))
	}
	if title != "" {
		args = append(args, "-T", title)
	}
	args = append(args, command)
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// BindKey binds a key to a command for a session.
func BindKey(session, key, command string) error {
	// Use bind-key with -T root for global bindings (like mouse events)
	cmd := exec.Command("tmux", "bind-key", "-T", "root", key, "run-shell", command)
	return cmd.Run()
}

// BindKeyPopup binds a key to display a command in a popup.
func BindKeyPopup(session, key, command string, width, height int, title, workingDir string) error {
	args := []string{"bind-key", "-T", "root", key, "display-popup", "-E"}
	if workingDir != "" {
		args = append(args, "-d", workingDir)
	}
	if width > 0 {
		args = append(args, "-w", strconv.Itoa(width))
	}
	if height > 0 {
		args = append(args, "-h", strconv.Itoa(height))
	}
	if title != "" {
		args = append(args, "-T", title)
	}
	args = append(args, command)
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// MenuItem represents an item in a tmux context menu.
type MenuItem struct {
	Label   string // Display text
	Key     string // Shortcut key (single char, or "" for none)
	Command string // tmux command to execute
}

// BindContextMenu binds a key to display a context menu.
// Uses -x M -y M to position at mouse coordinates, -O to keep menu open.
func BindContextMenu(key, title string, items []MenuItem) error {
	args := []string{"bind-key", "-T", "root", key, "display-menu", "-O", "-T", title, "-x", "M", "-y", "M"}
	for _, item := range items {
		args = append(args, item.Label, item.Key, item.Command)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// ApplyGlobalBindings sets up ORC's global tmux key bindings.
// Safe to call repeatedly (idempotent). Silently ignores errors (tmux may not be running).
func ApplyGlobalBindings() {
	// Session browser (prefix+s) with ORC context format
	// Shows: "Workshop Name [WORK-xxx] - Commission Title [COMM-xxx], ..."
	_ = exec.Command("tmux", "bind-key", "-T", "prefix", "s",
		"choose-tree", "-sZ", "-F",
		`#{session_name} [#{ORC_WORKSHOP_ID}] - #{?#{ORC_CONTEXT},#{ORC_CONTEXT},(idle)}`).Run()

	// ORC session picker (prefix+S) with rich agent/focus display and preview
	// Uses display-popup to run TUI with split preview pane
	_ = exec.Command("tmux", "bind-key", "-T", "prefix", "S",
		"display-popup", "-E", "-w", "90%", "-h", "90%", "$HOME/.orc/tmux/orc-session-picker.sh").Run()

	// Double-click status bar → orc summary popup
	_ = BindKeyPopup("", "DoubleClick1Status",
		"CLICOLOR_FORCE=1 orc summary | less -R -X",
		100, 30, "ORC Summary", "#{pane_current_path}")

	// Right-click status bar → context menu
	_ = BindContextMenu("MouseDown3Status", " ORC ", []MenuItem{
		// ORC custom options
		{Label: "Show Summary", Key: "s", Command: "display-popup -E -w 100 -h 30 -T 'ORC Summary' 'cd #{pane_current_path} && CLICOLOR_FORCE=1 orc summary | less -R -X'"},
		// Separator
		{Label: "", Key: "", Command: ""},
		// Default tmux window options
		{Label: "Swap Left", Key: "<", Command: "swap-window -t :-1"},
		{Label: "Swap Right", Key: ">", Command: "swap-window -t :+1"},
		{Label: "#{?pane_marked,Unmark,Mark}", Key: "m", Command: "select-pane -m"},
		{Label: "Kill", Key: "X", Command: "kill-window"},
		{Label: "Respawn", Key: "R", Command: "respawn-window -k"},
		{Label: "Rename", Key: "r", Command: "command-prompt -I \"#W\" \"rename-window -- '%%'\""},
		{Label: "New Window", Key: "c", Command: "new-window"},
	})
}

// SetEnvironment sets an environment variable for a tmux session.
func SetEnvironment(sessionName, key, value string) error {
	cmd := exec.Command("tmux", "set-environment", "-t", exactSession(sessionName), key, value)
	return cmd.Run()
}

// GetEnvironment gets an environment variable from a tmux session.
// Returns the value, or error if not found.
func GetEnvironment(sessionName, key string) (string, error) {
	cmd := exec.Command("tmux", "show-environment", "-t", exactSession(sessionName), key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Output format: "KEY=value\n"
	line := strings.TrimSpace(string(output))
	if strings.HasPrefix(line, key+"=") {
		return strings.TrimPrefix(line, key+"="), nil
	}
	return "", fmt.Errorf("env var %s not found", key)
}

// ListSessions returns all tmux session names.
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var sessions []string
	for _, line := range lines {
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}

// FindSessionByWorkshopID finds the session with ORC_WORKSHOP_ID=workshopID.
// Returns session name, or empty string if not found.
func FindSessionByWorkshopID(workshopID string) string {
	sessions, err := ListSessions()
	if err != nil {
		return ""
	}
	for _, session := range sessions {
		val, err := GetEnvironment(session, "ORC_WORKSHOP_ID")
		if err == nil && val == workshopID {
			return session
		}
	}
	return ""
}

// GetWindowOption gets a window option value.
// target format: "session:window" (e.g., "mysession:1" or "mysession:mywindow")
func GetWindowOption(target, option string) string {
	cmd := exec.Command("tmux", "show-options", "-t", target, "-wqv", option)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// SetWindowOption sets a window option value.
// target format: "session:window" (e.g., "mysession:1" or "mysession:mywindow")
func SetWindowOption(target, option, value string) error {
	cmd := exec.Command("tmux", "set-option", "-t", target, "-w", option, value)
	return cmd.Run()
}

// ListWindows returns window names in a session.
func ListWindows(sessionName string) ([]string, error) {
	cmd := exec.Command("tmux", "list-windows", "-t", exactSession(sessionName), "-F", "#{window_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var windows []string
	for _, line := range lines {
		if line != "" {
			windows = append(windows, line)
		}
	}
	return windows, nil
}

// PaneInfo contains information about a tmux pane
type PaneInfo struct {
	ID        string // e.g., "%0", "%1"
	Index     int    // pane index in window
	HasRole   bool   // whether @pane_role tmux option is set
	RoleValue string // value of @pane_role if set
}

// ListPanes returns information about all panes in a window
func ListPanes(sessionName, windowName string) ([]PaneInfo, error) {
	target := exactTarget(sessionName, windowName)

	// Get pane IDs and indices
	cmd := exec.Command("tmux", "list-panes", "-t", target, "-F", "#{pane_id}:#{pane_index}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	panes := make([]PaneInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		paneID := parts[0]
		paneIndex, _ := strconv.Atoi(parts[1])

		// Check if @pane_role option is set (tmux pane option, set by gotmux adapter)
		// Note: we use @pane_role (tmux option) NOT PANE_ROLE (shell env var) because
		// tmux format strings cannot read shell environment variables.
		roleCmd := exec.Command("tmux", "display-message", "-t", paneID, "-p", "#{@pane_role}")
		roleOutput, _ := roleCmd.Output()
		role := strings.TrimSpace(string(roleOutput))

		paneInfo := PaneInfo{
			ID:        paneID,
			Index:     paneIndex,
			HasRole:   role != "",
			RoleValue: role,
		}

		panes = append(panes, paneInfo)
	}

	return panes, nil
}

// BreakPane breaks a pane into a new window
func BreakPane(paneID, targetWindow string) error {
	cmd := exec.Command("tmux", "break-pane", "-s", paneID, "-t", targetWindow)
	return cmd.Run()
}

// MoveWindowAfter moves a window to be positioned after another window.
// If the window is already at afterIndex+1, the move is skipped.
// If the target index is occupied, the error is returned to the caller.
func MoveWindowAfter(sessionName, windowName, afterWindow string) error {
	// Get the index of the afterWindow
	afterTarget := exactTarget(sessionName, afterWindow)
	cmd := exec.Command("tmux", "display-message", "-t", afterTarget, "-p", "#{window_index}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get window index for %s: %w", afterWindow, err)
	}

	afterIndex, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse window index: %w", err)
	}

	// Get the current index of the window being moved
	moveTarget := exactTarget(sessionName, windowName)
	cmd = exec.Command("tmux", "display-message", "-t", moveTarget, "-p", "#{window_index}")
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get window index for %s: %w", windowName, err)
	}

	currentIndex, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse current window index: %w", err)
	}

	// Already in the correct position — nothing to do
	newIndex := afterIndex + 1
	if currentIndex == newIndex {
		return nil
	}

	cmd = exec.Command("tmux", "move-window", "-s", moveTarget, "-t", fmt.Sprintf("=%s:%d", sessionName, newIndex))
	return cmd.Run()
}

// RefreshWorkbenchLayout relocates guest panes to a sibling -imps window
func RefreshWorkbenchLayout(sessionName, workbenchWindow string) error {
	// 1. List all panes in the workbench window
	panes, err := ListPanes(sessionName, workbenchWindow)
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}

	// 2. Find guest panes (no PANE_ROLE)
	var guestPanes []PaneInfo
	for _, pane := range panes {
		if !pane.HasRole {
			guestPanes = append(guestPanes, pane)
		}
	}

	// If no guests, nothing to do
	if len(guestPanes) == 0 {
		return nil
	}

	// 3. Create or find the -imps window
	impsWindow := workbenchWindow + "-imps"
	impsExists := WindowExists(sessionName, impsWindow)

	if !impsExists {
		// Create new window after the workbench window
		// Break the first guest pane into a new window in the correct session.
		// Without -t, break-pane creates the window in the CALLER's session, not the source pane's session.
		firstGuest := guestPanes[0]
		sessionTarget := exactSession(sessionName) + ":"
		cmd := exec.Command("tmux", "break-pane", "-s", firstGuest.ID, "-t", sessionTarget, "-n", impsWindow)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create imps window: %w", err)
		}

		// Move it to be right after the workbench window
		if err := MoveWindowAfter(sessionName, impsWindow, workbenchWindow); err != nil {
			// Non-fatal - window was created, just not in ideal position
			fmt.Printf("Warning: failed to position %s window: %v\n", impsWindow, err)
		}

		// Remove first guest from list (already relocated)
		guestPanes = guestPanes[1:]
	}

	// 4. Relocate remaining guest panes to the imps window
	impsTarget := exactTarget(sessionName, impsWindow)
	for _, guest := range guestPanes {
		// Use join-pane to move guest to imps window
		if err := JoinPane(guest.ID, impsTarget, false, 0); err != nil {
			fmt.Printf("Warning: failed to relocate pane %s: %v\n", guest.ID, err)
		}
	}

	return nil
}

// SetPaneTitle sets the title of a pane using select-pane -T
func SetPaneTitle(paneID, title string) error {
	cmd := exec.Command("tmux", "select-pane", "-t", paneID, "-T", title)
	return cmd.Run()
}

// EnrichSession applies ORC enrichment to all windows in a session
// This includes setting pane titles and window options (NOT PANE_ROLE - that must be set at pane creation)
func EnrichSession(sessionName string) error {
	// Get all windows in the session
	windows, err := ListWindows(sessionName)
	if err != nil {
		return fmt.Errorf("failed to list windows: %w", err)
	}

	// Process each window
	for _, window := range windows {
		if err := enrichWindow(sessionName, window); err != nil {
			// Log warning but continue with other windows
			fmt.Printf("Warning: failed to enrich window %s: %v\n", window, err)
		}
	}

	return nil
}

// enrichWindow applies enrichment to a single window
func enrichWindow(sessionName, windowName string) error {
	// Get all panes in the window
	panes, err := ListPanes(sessionName, windowName)
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}

	// For workbench windows: pane 0=vim, pane 1=goblin, pane 2=shell
	// For -imps windows: all panes are guests (skip enrichment)
	isImpsWindow := strings.HasSuffix(windowName, "-imps")
	if isImpsWindow {
		return nil
	}

	for _, pane := range panes {
		// Determine title based on PANE_ROLE if set, otherwise use index heuristic
		var title string
		if pane.HasRole {
			title = fmt.Sprintf("%s [%s]", pane.RoleValue, windowName)
		} else {
			// Use index heuristic for pane title (gotmux standard layout)
			switch pane.Index {
			case 0:
				title = fmt.Sprintf("vim [%s]", windowName)
			case 1:
				title = fmt.Sprintf("goblin [%s]", windowName)
			case 2:
				title = fmt.Sprintf("shell [%s]", windowName)
			default:
				// Unknown pane index, skip
				continue
			}
		}

		// Set pane title (cosmetic - doesn't require shell access)
		_ = SetPaneTitle(pane.ID, title)
	}

	// Set window option @orc_enriched=1 to mark as enriched
	target := exactTarget(sessionName, windowName)
	_ = SetWindowOption(target, "@orc_enriched", "1")

	return nil
}
