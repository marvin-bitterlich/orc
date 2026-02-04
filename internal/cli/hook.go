package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/wire"
)

// HookCmd returns the hook command - parent for Claude Code hook handlers
func HookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook <event>",
		Short: "Handle Claude Code hook events",
		Long: `Process Claude Code hook events.

This command is called by Claude Code hooks and reads event data from stdin.
Each event has a specific handler subcommand.

Available events:
  Stop    - Called when Claude wants to stop the session

Example:
  echo '{"session_id":"abc"}' | orc hook Stop`,
	}

	// Add event handlers as subcommands
	cmd.AddCommand(hookStopCmd())

	return cmd
}

// StopHookEvent represents the JSON payload from Claude Code Stop hook
type StopHookEvent struct {
	StopHookActive bool   `json:"stop_hook_active"`
	Cwd            string `json:"cwd"`
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
}

// StopHookResponse represents the JSON response to block a stop
type StopHookResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// hookStopCmd handles the Stop event for IMP context
func hookStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "Stop",
		Short: "Handle Stop event (IMP context)",
		Long:  "Called when Claude wants to stop. Blocks if IMP has incomplete work.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookStop()
		},
	}
}

func runHookStop() error {
	ctx := context.Background()

	// 1. Read stdin JSON
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		// Can't read stdin - allow stop (fail open)
		return nil //nolint:nilerr // intentional fail-open design
	}

	// 2. Parse hook event
	var event StopHookEvent
	if err := json.Unmarshal(data, &event); err != nil {
		// Invalid JSON - allow stop (fail open)
		return nil //nolint:nilerr // intentional fail-open design
	}

	// 3. Check stop_hook_active first (prevent infinite loop)
	if event.StopHookActive {
		return nil
	}

	// 4. Get cwd from event - this is where Claude Code session is running
	cwd := event.Cwd
	if cwd == "" {
		// No cwd provided - allow stop
		return nil
	}

	// 5. Load .orc/config.json from cwd
	cfg, err := config.LoadConfig(cwd)
	if err != nil {
		// No config - allow stop (not ORC context)
		return nil //nolint:nilerr // intentional fail-open design
	}

	// 6. Check if this is a workbench (BENCH-xxx)
	if !config.IsWorkbench(cfg.PlaceID) {
		// Not a workbench - allow stop
		return nil
	}

	workbenchID := cfg.PlaceID

	// 7. Get focused shipment for this workbench
	focusID, err := wire.WorkbenchService().GetFocusedID(ctx, workbenchID)
	if err != nil || focusID == "" {
		// No focus set - allow stop
		return nil //nolint:nilerr // intentional fail-open design
	}

	// 8. Check if focus is a shipment (SHIP-xxx)
	if !strings.HasPrefix(focusID, "SHIP-") {
		// Not focused on a shipment - allow stop
		return nil
	}

	// 8a. Get shipment to check status
	shipment, err := wire.ShipmentService().GetShipment(ctx, focusID)
	if err != nil {
		// Can't get shipment - allow stop (fail open)
		return nil //nolint:nilerr // intentional fail-open design
	}

	// 8b. Only block in auto_implementing mode - implementing (manual) mode allows stop
	if shipment.Status != "auto_implementing" {
		return nil
	}

	// 9. Get tasks for the focused shipment
	tasks, err := wire.ShipmentService().GetShipmentTasks(ctx, focusID)
	if err != nil {
		// Can't get tasks - allow stop (fail open)
		return nil //nolint:nilerr // intentional fail-open design
	}

	// 10. Count incomplete tasks
	incompleteCount := 0
	for _, task := range tasks {
		if task.Status != "complete" {
			incompleteCount++
		}
	}

	// 11. If no incomplete tasks, allow stop
	if incompleteCount == 0 {
		return nil
	}

	// 12. Block with workflow guidance
	response := StopHookResponse{
		Decision: "block",
		Reason: `STOP. You are an IMP operating autonomously in an orchestration system. Your focused shipment has incomplete tasks. You do not stop until the shipment is complete.

WORKFLOW:
1. No plan? Run /imp-plan-create
2. Plan exists? Implement it, then run /imp-rec
3. Stuck? Run /imp-escalate
4. Task complete? Run /imp-rec to chain to next

DO NOT STOP. Execute the appropriate /imp- command now.`,
	}

	// Output JSON response (exit 0 with output = block)
	output, _ := json.Marshal(response)
	fmt.Fprintln(os.Stdout, string(output))

	return nil
}
