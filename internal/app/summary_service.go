package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
)

// SummaryServiceImpl implements the SummaryService interface.
type SummaryServiceImpl struct {
	commissionService primary.CommissionService
	tomeService       primary.TomeService
	shipmentService   primary.ShipmentService
	taskService       primary.TaskService
	noteService       primary.NoteService
	workbenchService  primary.WorkbenchService
	planService       primary.PlanService
	approvalService   primary.ApprovalService
	escalationService primary.EscalationService
	receiptService    primary.ReceiptService
}

// NewSummaryService creates a new SummaryService with injected dependencies.
func NewSummaryService(
	commissionService primary.CommissionService,
	tomeService primary.TomeService,
	shipmentService primary.ShipmentService,
	taskService primary.TaskService,
	noteService primary.NoteService,
	workbenchService primary.WorkbenchService,
	planService primary.PlanService,
	approvalService primary.ApprovalService,
	escalationService primary.EscalationService,
	receiptService primary.ReceiptService,
) *SummaryServiceImpl {
	return &SummaryServiceImpl{
		commissionService: commissionService,
		tomeService:       tomeService,
		shipmentService:   shipmentService,
		taskService:       taskService,
		noteService:       noteService,
		workbenchService:  workbenchService,
		planService:       planService,
		approvalService:   approvalService,
		escalationService: escalationService,
		receiptService:    receiptService,
	}
}

// GetCommissionSummary returns a flat summary of shipments and tomes under a commission.
func (s *SummaryServiceImpl) GetCommissionSummary(ctx context.Context, req primary.SummaryRequest) (*primary.CommissionSummary, error) {
	// Debug helper
	var debugMsgs []string
	addDebug := func(msg string) {
		if req.DebugMode {
			debugMsgs = append(debugMsgs, msg)
		}
	}

	// Validate commission exists
	commission, err := s.commissionService.GetCommission(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commission: %w", err)
	}

	// Fetch all tomes for this commission
	allTomes, err := s.tomeService.ListTomes(ctx, primary.TomeFilters{CommissionID: req.CommissionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list tomes: %w", err)
	}

	// Fetch all shipments for this commission
	allShipments, err := s.shipmentService.ListShipments(ctx, primary.ShipmentFilters{CommissionID: req.CommissionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}

	addDebug(fmt.Sprintf("Fetched %d tomes, %d shipments", len(allTomes), len(allShipments)))

	// Build flat shipment list
	var shipmentSummaries []primary.ShipmentSummary
	for _, ship := range allShipments {
		if ship.Status == "complete" {
			addDebug(fmt.Sprintf("Hidden: %s (%s) - status is complete", ship.ID, ship.Title))
			continue
		}

		shipSummary, err := s.buildShipmentSummary(ctx, ship, req.FocusID)
		if err != nil {
			continue // Skip on error
		}
		shipmentSummaries = append(shipmentSummaries, *shipSummary)
	}

	// Build flat tome list
	var tomeSummaries []primary.TomeSummary
	for _, tome := range allTomes {
		if tome.Status == "closed" {
			addDebug(fmt.Sprintf("Hidden: %s (%s) - status is closed", tome.ID, tome.Title))
			continue
		}
		expandNotes := tome.ID == req.FocusID
		tomeSummary, err := s.buildTomeSummary(ctx, tome, req.FocusID, expandNotes)
		if err != nil {
			continue // Skip on error
		}
		tomeSummaries = append(tomeSummaries, *tomeSummary)
	}

	// Determine if this is the focused commission
	isFocusedCommission := false
	if req.FocusID != "" {
		if req.FocusID == commission.ID {
			isFocusedCommission = true
		}
		for _, tome := range allTomes {
			if tome.ID == req.FocusID {
				isFocusedCommission = true
				break
			}
		}
		for _, ship := range allShipments {
			if ship.ID == req.FocusID {
				isFocusedCommission = true
				break
			}
		}
	}

	// Build debug info if in debug mode
	var debugInfo *primary.DebugInfo
	if req.DebugMode && len(debugMsgs) > 0 {
		debugInfo = &primary.DebugInfo{Messages: debugMsgs}
	}

	return &primary.CommissionSummary{
		ID:                  commission.ID,
		Title:               commission.Title,
		IsFocusedCommission: isFocusedCommission,
		Shipments:           shipmentSummaries,
		Tomes:               tomeSummaries,
		DebugInfo:           debugInfo,
	}, nil
}

// buildTomeSummary creates a TomeSummary with note count.
// When expandNotes is true, includes the full Notes slice (for focused tomes/conclaves).
func (s *SummaryServiceImpl) buildTomeSummary(ctx context.Context, tome *primary.Tome, focusID string, expandNotes bool) (*primary.TomeSummary, error) {
	// Get notes for this tome
	notes, err := s.tomeService.GetTomeNotes(ctx, tome.ID)
	noteCount := 0
	var noteSummaries []primary.NoteSummary

	if err == nil {
		for _, n := range notes {
			if n.Status != "closed" {
				noteCount++
				if expandNotes {
					noteSummaries = append(noteSummaries, primary.NoteSummary{
						ID:    n.ID,
						Title: n.Title,
						Type:  n.Type,
					})
				}
			}
		}
	}

	return &primary.TomeSummary{
		ID:        tome.ID,
		Title:     tome.Title,
		Status:    tome.Status,
		NoteCount: noteCount,
		IsFocused: tome.ID == focusID,
		Pinned:    tome.Pinned,
		Notes:     noteSummaries,
	}, nil
}

// buildShipmentSummary creates a ShipmentSummary with task progress.
func (s *SummaryServiceImpl) buildShipmentSummary(ctx context.Context, ship *primary.Shipment, focusID string) (*primary.ShipmentSummary, error) {
	// Get tasks for this shipment
	tasks, err := s.shipmentService.GetShipmentTasks(ctx, ship.ID)
	tasksDone := 0
	tasksTotal := 0
	var taskSummaries []primary.TaskSummary

	isFocused := ship.ID == focusID

	if err == nil {
		for _, t := range tasks {
			tasksTotal++
			if t.Status == "complete" {
				tasksDone++
			}
			// Include non-complete tasks for focused shipment
			if isFocused && t.Status != "complete" {
				taskSummary := primary.TaskSummary{
					ID:     t.ID,
					Title:  t.Title,
					Status: t.Status,
				}
				// Fetch children for focused shipment tasks
				s.fetchTaskChildren(ctx, &taskSummary)
				taskSummaries = append(taskSummaries, taskSummary)
			}
		}
	}

	// Get workbench name if assigned
	benchName := ""
	if ship.AssignedWorkbenchID != "" {
		bench, err := s.workbenchService.GetWorkbench(ctx, ship.AssignedWorkbenchID)
		if err == nil && bench != nil {
			benchName = bench.Name
		}
	}

	return &primary.ShipmentSummary{
		ID:         ship.ID,
		Title:      ship.Title,
		Status:     ship.Status,
		IsFocused:  isFocused,
		Pinned:     ship.Pinned,
		BenchID:    ship.AssignedWorkbenchID,
		BenchName:  benchName,
		TasksDone:  tasksDone,
		TasksTotal: tasksTotal,
		Tasks:      taskSummaries,
	}, nil
}

// fetchTaskChildren populates the Plans, Approvals, Escalations, and Receipts for a task.
func (s *SummaryServiceImpl) fetchTaskChildren(ctx context.Context, task *primary.TaskSummary) {
	// Fetch plans for this task
	if s.planService != nil {
		plans, err := s.planService.ListPlans(ctx, primary.PlanFilters{TaskID: task.ID})
		if err == nil {
			for _, p := range plans {
				task.Plans = append(task.Plans, primary.PlanSummary{
					ID:     p.ID,
					Status: p.Status,
				})
			}
		}
	}

	// Fetch approvals for this task
	if s.approvalService != nil {
		approvals, err := s.approvalService.ListApprovals(ctx, primary.ApprovalFilters{TaskID: task.ID})
		if err == nil {
			for _, a := range approvals {
				task.Approvals = append(task.Approvals, primary.ApprovalSummary{
					ID:      a.ID,
					Outcome: a.Outcome,
				})
			}
		}
	}

	// Fetch escalations for this task
	if s.escalationService != nil {
		escalations, err := s.escalationService.ListEscalations(ctx, primary.EscalationFilters{TaskID: task.ID})
		if err == nil {
			for _, e := range escalations {
				task.Escalations = append(task.Escalations, primary.EscalationSummary{
					ID:            e.ID,
					Status:        e.Status,
					TargetActorID: e.TargetActorID,
				})
			}
		}
	}

	// Fetch receipts for this task
	if s.receiptService != nil {
		receipts, err := s.receiptService.ListReceipts(ctx, primary.ReceiptFilters{TaskID: task.ID})
		if err == nil {
			for _, r := range receipts {
				task.Receipts = append(task.Receipts, primary.ReceiptSummary{
					ID:     r.ID,
					Status: r.Status,
				})
			}
		}
	}
}

// Ensure SummaryServiceImpl implements the interface
var _ primary.SummaryService = (*SummaryServiceImpl)(nil)
