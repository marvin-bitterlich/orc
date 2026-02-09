package cli

import (
	gocontext "context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/ports/primary"
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
  Stop              - Called when Claude wants to stop the session
  UserPromptSubmit  - Called when user submits a prompt

Example:
  echo '{"session_id":"abc"}' | orc hook Stop`,
	}

	// Add event handlers as subcommands
	cmd.AddCommand(hookStopCmd())
	cmd.AddCommand(hookUserPromptSubmitCmd())

	// Add event viewing commands
	cmd.AddCommand(hookTailCmd())
	cmd.AddCommand(hookShowCmd())

	return cmd
}

// StopHookEvent represents the JSON payload from Claude Code Stop hook
type StopHookEvent struct {
	StopHookActive bool   `json:"stop_hook_active"`
	Cwd            string `json:"cwd"`
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
}

// UserPromptSubmitHookEvent represents the JSON payload from Claude Code UserPromptSubmit hook
type UserPromptSubmitHookEvent struct {
	Cwd            string `json:"cwd"`
	SessionID      string `json:"session_id"`
	Prompt         string `json:"prompt"`
	TranscriptPath string `json:"transcript_path"`
}

// StopHookResponse represents the JSON response to block a stop
type StopHookResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// hookContext holds ORC context discovered during hook processing
type hookContext struct {
	workbenchID     string
	shipmentID      string
	shipmentStatus  string
	incompleteCount int
}

// lookupORCContext discovers ORC context from a working directory
func lookupORCContext(ctx gocontext.Context, cwd string) *hookContext {
	hctx := &hookContext{incompleteCount: -1}

	if cwd == "" {
		return hctx
	}

	// Load .orc/config.json from cwd
	cfg, err := config.LoadConfig(cwd)
	if err != nil {
		return hctx
	}

	// Check if this is a workbench (BENCH-xxx)
	if !config.IsWorkbench(cfg.PlaceID) {
		return hctx
	}
	hctx.workbenchID = cfg.PlaceID

	// Get focused shipment for this workbench
	focusID, err := wire.WorkbenchService().GetFocusedID(ctx, hctx.workbenchID)
	if err != nil || focusID == "" {
		return hctx
	}

	// Check if focus is a shipment (SHIP-xxx)
	if !strings.HasPrefix(focusID, "SHIP-") {
		return hctx
	}
	hctx.shipmentID = focusID

	// Get shipment to check status
	shipment, err := wire.ShipmentService().GetShipment(ctx, focusID)
	if err != nil {
		return hctx
	}
	hctx.shipmentStatus = shipment.Status

	// Get tasks for the focused shipment
	tasks, err := wire.ShipmentService().GetShipmentTasks(ctx, focusID)
	if err != nil {
		return hctx
	}

	// Count incomplete tasks
	hctx.incompleteCount = 0
	for _, task := range tasks {
		if task.Status != "complete" {
			hctx.incompleteCount++
		}
	}

	return hctx
}

// logHookEvent persists a hook event (best effort - errors are logged, not returned)
func logHookEvent(ctx gocontext.Context, req primary.LogHookEventRequest) {
	_, err := wire.HookEventService().LogHookEvent(ctx, req)
	if err != nil {
		// Log error but don't fail the hook (fail-open)
		fmt.Fprintf(os.Stderr, "orc: failed to log hook event: %v\n", err)
	}
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
	ctx := NewContext()
	startTime := time.Now()

	// Initialize event request (will be persisted at the end)
	eventReq := primary.LogHookEventRequest{
		HookType:            primary.HookTypeStop,
		Decision:            primary.HookDecisionAllow,
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}

	// Defer event logging to capture final state
	defer func() {
		eventReq.DurationMs = int(time.Since(startTime).Milliseconds())
		logHookEvent(ctx, eventReq)
	}()

	// 1. Read stdin JSON
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		eventReq.Error = fmt.Sprintf("failed to read stdin: %v", err)
		return nil //nolint:nilerr // intentional fail-open design
	}
	eventReq.PayloadJSON = string(data)

	// 2. Parse hook event
	var event StopHookEvent
	if err := json.Unmarshal(data, &event); err != nil {
		eventReq.Error = fmt.Sprintf("failed to parse JSON: %v", err)
		return nil //nolint:nilerr // intentional fail-open design
	}

	eventReq.Cwd = event.Cwd
	eventReq.SessionID = event.SessionID

	// 3. Check stop_hook_active first (prevent infinite loop)
	if event.StopHookActive {
		eventReq.Reason = "stop_hook_active=true (preventing loop)"
		return nil
	}

	// 4. Look up ORC context
	hctx := lookupORCContext(ctx, event.Cwd)
	eventReq.WorkbenchID = hctx.workbenchID
	eventReq.ShipmentID = hctx.shipmentID
	eventReq.ShipmentStatus = hctx.shipmentStatus
	eventReq.TaskCountIncomplete = hctx.incompleteCount

	// 5. Check if we have ORC context
	if hctx.workbenchID == "" {
		eventReq.Reason = "no workbench context"
		return nil
	}

	if hctx.shipmentID == "" {
		eventReq.Reason = "no shipment focused"
		return nil
	}

	// 6. Only block in auto_implementing mode
	if hctx.shipmentStatus != "auto_implementing" {
		eventReq.Reason = fmt.Sprintf("shipment status is %s (not auto_implementing)", hctx.shipmentStatus)
		return nil
	}

	// 7. If no incomplete tasks, allow stop
	if hctx.incompleteCount == 0 {
		eventReq.Reason = "all tasks complete"
		return nil
	}

	// 8. Block with workflow guidance
	eventReq.Decision = primary.HookDecisionBlock
	eventReq.Reason = fmt.Sprintf("%d incomplete tasks", hctx.incompleteCount)

	response := StopHookResponse{
		Decision: "block",
		Reason: `STOP. You are an IMP operating autonomously in an orchestration system. Your focused shipment has incomplete tasks. You do not stop until the shipment is complete.

WORKFLOW:
1. No plan? Run /imp-plan-create
2. Plan created? Run /imp-plan-submit to approve
3. Plan approved? Run /imp-implement to see plan and code
4. Implementation done? Run /imp-rec to verify and chain to next
5. Stuck? Run /imp-escalate

DO NOT STOP. Execute the appropriate /imp- command now.`,
	}

	// Output JSON response (exit 0 with output = block)
	output, _ := json.Marshal(response)
	fmt.Fprintln(os.Stdout, string(output))

	return nil
}

// hookUserPromptSubmitCmd handles the UserPromptSubmit event
func hookUserPromptSubmitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "UserPromptSubmit",
		Short: "Handle UserPromptSubmit event",
		Long:  "Called when user submits a prompt. Logs the event for tracking.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookUserPromptSubmit()
		},
	}
}

func runHookUserPromptSubmit() error {
	ctx := NewContext()
	startTime := time.Now()

	// Initialize event request
	eventReq := primary.LogHookEventRequest{
		HookType:            primary.HookTypeUserPromptSubmit,
		Decision:            primary.HookDecisionAllow,
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}

	// Defer event logging
	defer func() {
		eventReq.DurationMs = int(time.Since(startTime).Milliseconds())
		logHookEvent(ctx, eventReq)
	}()

	// 1. Read stdin JSON
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		eventReq.Error = fmt.Sprintf("failed to read stdin: %v", err)
		return nil //nolint:nilerr // intentional fail-open design
	}
	eventReq.PayloadJSON = string(data)

	// 2. Parse hook event
	var event UserPromptSubmitHookEvent
	if err := json.Unmarshal(data, &event); err != nil {
		eventReq.Error = fmt.Sprintf("failed to parse JSON: %v", err)
		return nil //nolint:nilerr // intentional fail-open design
	}

	eventReq.Cwd = event.Cwd
	eventReq.SessionID = event.SessionID

	// 3. Look up ORC context
	hctx := lookupORCContext(ctx, event.Cwd)
	eventReq.WorkbenchID = hctx.workbenchID
	eventReq.ShipmentID = hctx.shipmentID
	eventReq.ShipmentStatus = hctx.shipmentStatus
	eventReq.TaskCountIncomplete = hctx.incompleteCount

	// UserPromptSubmit always allows - it just logs
	eventReq.Reason = "user prompt logged"

	return nil
}

// hookTailCmd shows recent hook events
func hookTailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Show recent hook events",
		Long:  "Show recent hook invocation events (default 50)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookTail(cmd)
		},
	}

	cmd.Flags().IntP("limit", "n", 50, "Number of events to show")
	cmd.Flags().StringP("workbench", "w", "", "Filter by workbench ID (auto-detects from cwd)")
	cmd.Flags().StringP("type", "t", "", "Filter by hook type (Stop, UserPromptSubmit)")
	cmd.Flags().BoolP("follow", "f", false, "Follow mode: poll for new events")

	return cmd
}

func runHookTail(cmd *cobra.Command) error {
	ctx := NewContext()

	limit, _ := cmd.Flags().GetInt("limit")
	workbenchID, _ := cmd.Flags().GetString("workbench")
	hookType, _ := cmd.Flags().GetString("type")
	follow, _ := cmd.Flags().GetBool("follow")

	// Auto-detect workbench from cwd if not specified
	if workbenchID == "" {
		cwd, _ := os.Getwd()
		cfg, err := config.LoadConfig(cwd)
		if err == nil && config.IsWorkbench(cfg.PlaceID) {
			workbenchID = cfg.PlaceID
		}
	}

	filters := primary.HookEventFilters{
		WorkbenchID: workbenchID,
		HookType:    hookType,
		Limit:       limit,
	}

	// Initial fetch
	events, err := wire.HookEventService().ListHookEvents(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to fetch hook events: %w", err)
	}

	printHookEvents(events)

	// If --follow, poll for new events
	if follow {
		var lastTimestamp string
		if len(events) > 0 {
			lastTimestamp = events[0].Timestamp
		}

		for {
			time.Sleep(1 * time.Second)

			newEvents, err := wire.HookEventService().ListHookEvents(ctx, filters)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching events: %v\n", err)
				continue
			}

			// Print only events newer than lastTimestamp
			for i := len(newEvents) - 1; i >= 0; i-- {
				event := newEvents[i]
				if lastTimestamp == "" || event.Timestamp > lastTimestamp {
					printHookEvent(event)
					if event.Timestamp > lastTimestamp {
						lastTimestamp = event.Timestamp
					}
				}
			}
		}
	}

	return nil
}

// hookShowCmd shows full details for a single hook event
func hookShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <event-id>",
		Short: "Show hook event details",
		Long:  "Show full details for a specific hook event including payload JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookShow(args[0])
		},
	}
}

func runHookShow(eventID string) error {
	ctx := NewContext()

	event, err := wire.HookEventService().GetHookEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get hook event: %w", err)
	}

	fmt.Printf("Hook Event: %s\n", event.ID)
	fmt.Printf("Type: %s\n", event.HookType)
	fmt.Printf("Timestamp: %s\n", formatHookTimestamp(event.Timestamp))
	fmt.Printf("Decision: %s\n", formatDecision(event.Decision))
	if event.Reason != "" {
		fmt.Printf("Reason: %s\n", event.Reason)
	}
	fmt.Println()

	fmt.Println("Context:")
	if event.WorkbenchID != "" {
		fmt.Printf("  Workbench: %s\n", event.WorkbenchID)
	}
	if event.ShipmentID != "" {
		fmt.Printf("  Shipment: %s (%s)\n", event.ShipmentID, event.ShipmentStatus)
	}
	if event.TaskCountIncomplete >= 0 {
		fmt.Printf("  Incomplete Tasks: %d\n", event.TaskCountIncomplete)
	}
	if event.SessionID != "" {
		fmt.Printf("  Session: %s\n", event.SessionID)
	}
	if event.Cwd != "" {
		fmt.Printf("  Cwd: %s\n", event.Cwd)
	}
	if event.DurationMs >= 0 {
		fmt.Printf("  Duration: %dms\n", event.DurationMs)
	}
	if event.Error != "" {
		fmt.Printf("  Error: %s\n", event.Error)
	}

	if event.PayloadJSON != "" {
		fmt.Println()
		fmt.Println("Payload JSON:")
		// Pretty print JSON
		var prettyJSON map[string]interface{}
		if err := json.Unmarshal([]byte(event.PayloadJSON), &prettyJSON); err == nil {
			formatted, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
			fmt.Printf("  %s\n", formatted)
		} else {
			fmt.Printf("  %s\n", event.PayloadJSON)
		}
	}

	return nil
}

func printHookEvents(events []*primary.HookEvent) {
	if len(events) == 0 {
		fmt.Println("No hook events found.")
		return
	}

	fmt.Printf("Found %d hook events:\n\n", len(events))

	// Print in reverse order (oldest first) for tail view
	for i := len(events) - 1; i >= 0; i-- {
		printHookEvent(events[i])
	}
}

func printHookEvent(event *primary.HookEvent) {
	// Format: timestamp | BENCH | hook_type | SHIP-xxx (status) | DECISION | reason
	shipmentInfo := "-"
	if event.ShipmentID != "" {
		shipmentInfo = fmt.Sprintf("%s (%s)", event.ShipmentID, event.ShipmentStatus)
	}

	workbenchInfo := event.WorkbenchID
	if workbenchInfo == "" {
		workbenchInfo = "-"
	}

	reason := event.Reason
	if len(reason) > 40 {
		reason = reason[:40] + "..."
	}

	fmt.Printf("%s | %-10s | %-16s | %-26s | %-5s | %s\n",
		formatHookTimestamp(event.Timestamp),
		workbenchInfo,
		event.HookType,
		shipmentInfo,
		strings.ToUpper(event.Decision),
		reason,
	)
}

func formatHookTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatDecision(decision string) string {
	if decision == "block" {
		return "BLOCK"
	}
	return "allow"
}
