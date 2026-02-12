package tmux

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/GianlucaP106/gotmux/gotmux"
)

// GotmuxAdapter wraps gotmux library for session lifecycle management
type GotmuxAdapter struct {
	tmux *gotmux.Tmux
}

// NewGotmuxAdapter creates a new gotmux adapter
func NewGotmuxAdapter() (*GotmuxAdapter, error) {
	tmux, err := gotmux.DefaultTmux()
	if err != nil {
		return nil, fmt.Errorf("failed to create tmux client: %w", err)
	}
	return &GotmuxAdapter{
		tmux: tmux,
	}, nil
}

// escapeShellCommand works around a gotmux quoting bug where ShellCommand is
// wrapped in single quotes (e.g. 'orc connect'). The shell interprets that as a
// single token, so multi-word commands fail with "command not found" (status 127).
// By replacing spaces with ' ' (close-quote, space, open-quote), gotmux's wrapping
// produces 'orc' 'connect' which the shell correctly parses as separate words.
func escapeShellCommand(cmd string) string {
	return strings.ReplaceAll(cmd, " ", "' '")
}

// CreateWorkbenchSession creates a tmux session with a 3-pane workbench window.
// Layout: vim (left) | goblin (top-right) / shell (bottom-right)
// Uses NewSession + AddWorkbenchWindow for the initial window.
func (g *GotmuxAdapter) CreateWorkbenchSession(sessionName, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	// Create session with plain shell (no ShellCommand — AddWorkbenchWindow handles pane setup)
	session, err := g.tmux.NewSession(&gotmux.SessionOptions{
		Name:           sessionName,
		StartDirectory: workbenchPath,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Get the auto-created first window and rename it
	windows, err := session.ListWindows()
	if err != nil {
		return fmt.Errorf("failed to list windows: %w", err)
	}
	if len(windows) == 0 {
		return fmt.Errorf("no windows found in new session")
	}
	firstWindow := windows[0]

	// Set up the 3-pane layout on the existing first window
	return g.setupWorkbenchPanes(firstWindow, workbenchName, workbenchPath, workbenchID, workshopID)
}

// AddWorkbenchWindow creates a new window on an existing session with a 3-pane workbench layout.
// Layout: vim (left) | goblin (top-right) / shell (bottom-right)
// Pane options (@pane_role, @bench_id, @workshop_id) are set on all three panes.
func (g *GotmuxAdapter) AddWorkbenchWindow(session *gotmux.Session, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	// Create new window (no ShellCommand available on NewWindowOptions)
	window, err := session.NewWindow(&gotmux.NewWindowOptions{
		WindowName:     workbenchName,
		StartDirectory: workbenchPath,
		DoNotAttach:    true,
	})
	if err != nil {
		return fmt.Errorf("failed to create window %s: %w", workbenchName, err)
	}

	return g.setupWorkbenchPanes(window, workbenchName, workbenchPath, workbenchID, workshopID)
}

// setupWorkbenchPanes configures a window with the standard 3-pane workbench layout.
// The window must already exist with at least one pane. It will be renamed to workbenchName
// and populated with: vim (left) | goblin (top-right) / shell (bottom-right).
func (g *GotmuxAdapter) setupWorkbenchPanes(window *gotmux.Window, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	// Rename window to workbench name
	if err := window.Rename(workbenchName); err != nil {
		return fmt.Errorf("failed to rename window: %w", err)
	}

	// Get first pane (starts as plain shell)
	panes, err := window.ListPanes()
	if err != nil || len(panes) == 0 {
		return fmt.Errorf("failed to get initial pane: %w", err)
	}
	vimPane := panes[0]

	// Make vim the root process of the first pane via respawn-pane -k
	// (NewWindowOptions doesn't support ShellCommand, so we respawn)
	if err := exec.Command("tmux", "respawn-pane", "-t", vimPane.Id, "-k", "vim").Run(); err != nil {
		return fmt.Errorf("failed to respawn vim pane: %w", err)
	}

	// Split vertically to create goblin pane (right side) with orc connect as root process
	if err := vimPane.SplitWindow(&gotmux.SplitWindowOptions{
		SplitDirection: gotmux.PaneSplitDirectionVertical,
		StartDirectory: workbenchPath,
		ShellCommand:   escapeShellCommand("orc connect"),
	}); err != nil {
		return fmt.Errorf("failed to split for goblin pane: %w", err)
	}

	// Get the goblin pane (second pane after split)
	panes, err = window.ListPanes()
	if err != nil || len(panes) < 2 {
		return fmt.Errorf("failed to get goblin pane after split: %w", err)
	}
	goblinPane := panes[1]

	// Split goblin pane horizontally to create shell pane (bottom-right)
	if err := goblinPane.SplitWindow(&gotmux.SplitWindowOptions{
		SplitDirection: gotmux.PaneSplitDirectionHorizontal,
		StartDirectory: workbenchPath,
	}); err != nil {
		return fmt.Errorf("failed to split for shell pane: %w", err)
	}

	// Get the shell pane (third pane after split)
	panes, err = window.ListPanes()
	if err != nil || len(panes) < 3 {
		return fmt.Errorf("failed to get shell pane after split: %w", err)
	}
	shellPane := panes[2]

	// Set main-pane-width BEFORE applying layout — tmux uses the current option
	// value at layout-selection time, so the option must be set first.
	if err := window.SetOption("main-pane-width", "50%"); err != nil {
		return fmt.Errorf("failed to set main-pane-width: %w", err)
	}

	// Apply main-vertical layout (workaround: use string, not constant — gotmux bug)
	if err := window.SelectLayout("main-vertical"); err != nil {
		return fmt.Errorf("failed to apply main-vertical layout: %w", err)
	}

	// Set tmux pane options for identity on all three panes
	// @pane_role is authoritative for pane identity (readable via #{@pane_role})
	// @bench_id and @workshop_id provide context without shell env vars
	for _, p := range []struct {
		pane *gotmux.Pane
		role string
	}{
		{vimPane, "vim"},
		{goblinPane, "goblin"},
		{shellPane, "shell"},
	} {
		if err := p.pane.SetOption("@pane_role", p.role); err != nil {
			return fmt.Errorf("failed to set @pane_role=%s: %w", p.role, err)
		}
		if err := p.pane.SetOption("@bench_id", workbenchID); err != nil {
			return fmt.Errorf("failed to set @bench_id on %s pane: %w", p.role, err)
		}
		if err := p.pane.SetOption("@workshop_id", workshopID); err != nil {
			return fmt.Errorf("failed to set @workshop_id on %s pane: %w", p.role, err)
		}
	}

	return nil
}

// GetSession returns a gotmux Session by name, or nil if not found.
func (g *GotmuxAdapter) GetSession(name string) (*gotmux.Session, error) {
	sessions, err := g.tmux.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	for _, s := range sessions {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, nil
}

// SessionExists checks if a tmux session exists
func (g *GotmuxAdapter) SessionExists(name string) bool {
	sessions, err := g.tmux.ListSessions()
	if err != nil {
		return false
	}
	for _, s := range sessions {
		if s.Name == name {
			return true
		}
	}
	return false
}

// KillSession terminates a tmux session
func (g *GotmuxAdapter) KillSession(name string) error {
	sessions, err := g.tmux.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}
	for _, s := range sessions {
		if s.Name == name {
			return s.Kill()
		}
	}
	return fmt.Errorf("session %s not found", name)
}

// ApplyActionType identifies the kind of reconciliation action.
type ApplyActionType string

const (
	ActionCreateSession   ApplyActionType = "CreateSession"
	ActionAddWindow       ApplyActionType = "AddWindow"
	ActionRelocateGuests  ApplyActionType = "RelocateGuests"
	ActionPruneDeadPanes  ApplyActionType = "PruneDeadPanes"
	ActionKillEmptyImps   ApplyActionType = "KillEmptyImpsWindow"
	ActionReconcileLayout ApplyActionType = "ReconcileLayout"
	ActionApplyEnrichment ApplyActionType = "ApplyEnrichment"
)

// ApplyAction represents a single reconciliation action in the plan.
type ApplyAction struct {
	Type        ApplyActionType
	Description string

	// Context fields used during execution (not all are set for every action type)
	SessionName   string
	WindowName    string
	WorkbenchName string
	WorkbenchPath string
	WorkbenchID   string
	WorkshopID    string
}

// DesiredWorkbench describes a workbench that should exist as a window.
type DesiredWorkbench struct {
	Name       string
	Path       string
	ID         string
	WorkshopID string
}

// ApplyPlan contains the full reconciliation plan.
type ApplyPlan struct {
	SessionName   string
	SessionExists bool
	Actions       []ApplyAction
	WindowSummary []WindowStatus
}

// WindowStatus summarizes a window's current state for display.
type WindowStatus struct {
	Name      string
	PaneCount int
	Healthy   bool
	DeadPanes int
	IsImps    bool
}

// PlanApply compares desired state (workbenches) to actual tmux state and returns actions.
func (g *GotmuxAdapter) PlanApply(sessionName string, workbenches []DesiredWorkbench) (*ApplyPlan, error) {
	plan := &ApplyPlan{SessionName: sessionName}

	// Check if session exists
	session, err := g.GetSession(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check session: %w", err)
	}

	if session == nil {
		// Session doesn't exist — need to create everything
		plan.SessionExists = false

		if len(workbenches) == 0 {
			return plan, nil
		}

		// First workbench creates the session
		first := workbenches[0]
		plan.Actions = append(plan.Actions, ApplyAction{
			Type:          ActionCreateSession,
			Description:   fmt.Sprintf("Create session %s with window %s", sessionName, first.Name),
			SessionName:   sessionName,
			WorkbenchName: first.Name,
			WorkbenchPath: first.Path,
			WorkbenchID:   first.ID,
			WorkshopID:    first.WorkshopID,
		})

		// Remaining workbenches get added as windows
		for _, wb := range workbenches[1:] {
			plan.Actions = append(plan.Actions, ApplyAction{
				Type:          ActionAddWindow,
				Description:   fmt.Sprintf("Add window %s (%s)", wb.Name, wb.ID),
				SessionName:   sessionName,
				WorkbenchName: wb.Name,
				WorkbenchPath: wb.Path,
				WorkbenchID:   wb.ID,
				WorkshopID:    wb.WorkshopID,
			})
		}

		// Always enrich at the end
		plan.Actions = append(plan.Actions, ApplyAction{
			Type:        ActionApplyEnrichment,
			Description: "Apply ORC enrichment (bindings, pane titles)",
			SessionName: sessionName,
		})

		return plan, nil
	}

	// Session exists — reconcile windows
	plan.SessionExists = true

	windows, err := session.ListWindows()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	// Build set of existing window names
	existingWindows := make(map[string]*gotmux.Window, len(windows))
	for _, w := range windows {
		existingWindows[w.Name] = w
	}

	// Check which workbenches need windows
	for _, wb := range workbenches {
		if _, exists := existingWindows[wb.Name]; !exists {
			plan.Actions = append(plan.Actions, ApplyAction{
				Type:          ActionAddWindow,
				Description:   fmt.Sprintf("Add window %s (%s)", wb.Name, wb.ID),
				SessionName:   sessionName,
				WorkbenchName: wb.Name,
				WorkbenchPath: wb.Path,
				WorkbenchID:   wb.ID,
				WorkshopID:    wb.WorkshopID,
			})
		}
	}

	// Check each window for health
	for _, w := range windows {
		isImps := strings.HasSuffix(w.Name, "-imps")

		panes, err := w.ListPanes()
		if err != nil {
			continue
		}

		deadCount := 0
		guestCount := 0
		for _, p := range panes {
			if p.Dead {
				deadCount++
			}
			// Check for guest panes (no @pane_role) in workbench windows
			if !isImps {
				opt, err := p.Option("@pane_role")
				if err != nil || opt == nil || opt.Value == "" {
					guestCount++
				}
			}
		}

		ws := WindowStatus{
			Name:      w.Name,
			PaneCount: len(panes),
			DeadPanes: deadCount,
			Healthy:   deadCount == 0 && (!isImps || len(panes) > 0),
			IsImps:    isImps,
		}
		plan.WindowSummary = append(plan.WindowSummary, ws)

		if isImps {
			// Check if all panes in -imps window are dead
			if deadCount > 0 && deadCount == len(panes) {
				plan.Actions = append(plan.Actions, ApplyAction{
					Type:        ActionKillEmptyImps,
					Description: fmt.Sprintf("Kill empty %s window (all %d panes dead)", w.Name, deadCount),
					SessionName: sessionName,
					WindowName:  w.Name,
				})
			} else if deadCount > 0 {
				plan.Actions = append(plan.Actions, ApplyAction{
					Type:        ActionPruneDeadPanes,
					Description: fmt.Sprintf("Prune %d dead panes in %s", deadCount, w.Name),
					SessionName: sessionName,
					WindowName:  w.Name,
				})
			}
		} else {
			// Workbench window — check for guest panes to relocate
			if guestCount > 0 {
				plan.Actions = append(plan.Actions, ApplyAction{
					Type:        ActionRelocateGuests,
					Description: fmt.Sprintf("Relocate %d guest panes from %s to %s-imps", guestCount, w.Name, w.Name),
					SessionName: sessionName,
					WindowName:  w.Name,
				})
			}
		}
	}

	// Always reconcile layout on all workbench windows
	for _, wb := range workbenches {
		if _, exists := existingWindows[wb.Name]; exists {
			plan.Actions = append(plan.Actions, ApplyAction{
				Type:        ActionReconcileLayout,
				Description: fmt.Sprintf("Reconcile layout on %s (main-pane-width 50%%)", wb.Name),
				SessionName: sessionName,
				WindowName:  wb.Name,
			})
		}
	}

	// Always enrich at the end
	plan.Actions = append(plan.Actions, ApplyAction{
		Type:        ActionApplyEnrichment,
		Description: "Apply ORC enrichment (bindings, pane titles)",
		SessionName: sessionName,
	})

	return plan, nil
}

// ExecutePlan executes all actions in a plan sequentially.
func (g *GotmuxAdapter) ExecutePlan(plan *ApplyPlan) error {
	for _, action := range plan.Actions {
		if err := g.executeAction(action); err != nil {
			return fmt.Errorf("action %s failed: %w", action.Type, err)
		}
	}
	return nil
}

// executeAction dispatches a single reconciliation action.
func (g *GotmuxAdapter) executeAction(action ApplyAction) error {
	switch action.Type {
	case ActionCreateSession:
		return g.CreateWorkbenchSession(action.SessionName, action.WorkbenchName, action.WorkbenchPath, action.WorkbenchID, action.WorkshopID)

	case ActionAddWindow:
		session, err := g.GetSession(action.SessionName)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		if session == nil {
			return fmt.Errorf("session %s not found", action.SessionName)
		}
		return g.AddWorkbenchWindow(session, action.WorkbenchName, action.WorkbenchPath, action.WorkbenchID, action.WorkshopID)

	case ActionRelocateGuests:
		return RefreshWorkbenchLayout(action.SessionName, action.WindowName)

	case ActionPruneDeadPanes:
		return g.pruneDeadPanes(action.SessionName, action.WindowName)

	case ActionKillEmptyImps:
		return KillWindow(action.SessionName, action.WindowName)

	case ActionReconcileLayout:
		return g.reconcileLayout(action.SessionName, action.WindowName)

	case ActionApplyEnrichment:
		ApplyGlobalBindings()
		return EnrichSession(action.SessionName)

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// pruneDeadPanes kills dead panes in a window.
func (g *GotmuxAdapter) pruneDeadPanes(sessionName, windowName string) error {
	session, err := g.GetSession(sessionName)
	if err != nil || session == nil {
		return fmt.Errorf("session %s not found", sessionName)
	}

	window, err := session.GetWindowByName(windowName)
	if err != nil || window == nil {
		return fmt.Errorf("window %s not found", windowName)
	}

	panes, err := window.ListPanes()
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}

	for _, p := range panes {
		if p.Dead {
			if err := p.Kill(); err != nil {
				return fmt.Errorf("failed to kill dead pane %s: %w", p.Id, err)
			}
		}
	}

	return nil
}

// reconcileLayout ensures a workbench window has the correct layout settings.
func (g *GotmuxAdapter) reconcileLayout(sessionName, windowName string) error {
	session, err := g.GetSession(sessionName)
	if err != nil || session == nil {
		return fmt.Errorf("session %s not found", sessionName)
	}

	window, err := session.GetWindowByName(windowName)
	if err != nil || window == nil {
		return fmt.Errorf("window %s not found", windowName)
	}

	// Set main-pane-width BEFORE applying layout — tmux uses the current option
	// value at layout-selection time, so the option must be set first.
	if err := window.SetOption("main-pane-width", "50%"); err != nil {
		return fmt.Errorf("failed to set main-pane-width: %w", err)
	}

	// Apply main-vertical layout
	if err := window.SelectLayout("main-vertical"); err != nil {
		return fmt.Errorf("failed to apply layout: %w", err)
	}

	return nil
}

// AttachInstructions returns instructions for attaching to a session
func (g *GotmuxAdapter) AttachInstructions(sessionName string) string {
	return fmt.Sprintf("Attach to session: tmux attach -t %s\n\n"+
		"Window Layout:\n"+
		"  ┌─────────────────────┬──────────────┐\n"+
		"  │                     │  goblin      │\n"+
		"  │      vim            ├──────────────┤\n"+
		"  │                     │  shell       │\n"+
		"  └─────────────────────┴──────────────┘\n\n"+
		"TMux Commands:\n"+
		"  Switch panes: Ctrl+b then arrow keys\n"+
		"  Detach session: Ctrl+b then d\n",
		sessionName)
}
