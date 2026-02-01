package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
)

// SummaryServiceImpl implements the SummaryService interface.
type SummaryServiceImpl struct {
	commissionService primary.CommissionService
	conclaveService   primary.ConclaveService
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
	conclaveService primary.ConclaveService,
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
		conclaveService:   conclaveService,
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

// GetCommissionSummary returns a hierarchical summary with role-based filtering.
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

	// 1. Fetch all conclaves for this commission
	conclaves, err := s.conclaveService.ListConclaves(ctx, primary.ConclaveFilters{CommissionID: req.CommissionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list conclaves: %w", err)
	}

	// 2. Fetch all tomes for this commission
	allTomes, err := s.tomeService.ListTomes(ctx, primary.TomeFilters{CommissionID: req.CommissionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list tomes: %w", err)
	}

	// 3. Fetch all shipments for this commission
	allShipments, err := s.shipmentService.ListShipments(ctx, primary.ShipmentFilters{CommissionID: req.CommissionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}

	addDebug(fmt.Sprintf("Fetched %d conclaves, %d tomes, %d shipments", len(conclaves), len(allTomes), len(allShipments)))

	// Build lookup maps for grouping
	tomesByContainer := s.groupTomesByContainer(allTomes)
	shipsByContainer := s.groupShipmentsByContainer(allShipments)

	// Build conclave summaries
	var conclaveSummaries []primary.ConclaveSummary
	for _, con := range conclaves {
		// Skip closed conclaves
		if con.Status == "closed" {
			addDebug(fmt.Sprintf("Hidden: %s (%s) - status is closed", con.ID, con.Title))
			continue
		}

		conSummary := primary.ConclaveSummary{
			ID:        con.ID,
			Title:     con.Title,
			Status:    con.Status,
			Pinned:    con.Pinned,
			IsFocused: con.ID == req.FocusID,
		}

		// Get tomes for this conclave
		conclaveIsFocused := con.ID == req.FocusID
		if tomes, ok := tomesByContainer[con.ID]; ok {
			for _, tome := range tomes {
				if tome.Status == "closed" {
					addDebug(fmt.Sprintf("Hidden: %s (%s) - status is closed", tome.ID, tome.Title))
					continue
				}
				// Expand notes if conclave is focused or tome itself is focused
				expandNotes := conclaveIsFocused || tome.ID == req.FocusID
				tomeSummary, err := s.buildTomeSummary(ctx, tome, req.FocusID, expandNotes)
				if err != nil {
					continue // Skip on error
				}
				conSummary.Tomes = append(conSummary.Tomes, *tomeSummary)
			}
		}

		// Get shipments for this conclave
		if ships, ok := shipsByContainer[con.ID]; ok {
			for _, ship := range ships {
				if ship.Status == "complete" {
					addDebug(fmt.Sprintf("Hidden: %s (%s) - status is complete", ship.ID, ship.Title))
					continue
				}

				shipSummary, err := s.buildShipmentSummary(ctx, ship, req.FocusID)
				if err != nil {
					continue // Skip on error
				}
				conSummary.Shipments = append(conSummary.Shipments, *shipSummary)
			}
		}

		conclaveSummaries = append(conclaveSummaries, conSummary)
	}

	// Build library summary (tomes with container_type="library")
	librarySummary := s.buildLibrarySummaryWithDebug(ctx, allTomes, req.ExpandLibrary, req.FocusID, addDebug)

	// Build shipyard summary (shipments with container_type="shipyard")
	shipyardSummary := s.buildShipyardSummaryWithDebug(ctx, allShipments, req.ExpandLibrary, req.FocusID, addDebug)

	// Build debug info if in debug mode
	var debugInfo *primary.DebugInfo
	if req.DebugMode && len(debugMsgs) > 0 {
		debugInfo = &primary.DebugInfo{Messages: debugMsgs}
	}

	return &primary.CommissionSummary{
		ID:        commission.ID,
		Title:     commission.Title,
		Conclaves: conclaveSummaries,
		Library:   librarySummary,
		Shipyard:  shipyardSummary,
		DebugInfo: debugInfo,
	}, nil
}

// groupTomesByContainer groups tomes by their container ID (conclave or library).
func (s *SummaryServiceImpl) groupTomesByContainer(tomes []*primary.Tome) map[string][]*primary.Tome {
	result := make(map[string][]*primary.Tome)
	for _, t := range tomes {
		containerID := t.ContainerID
		// Fall back to ConclaveID for backwards compatibility
		if containerID == "" && t.ConclaveID != "" {
			containerID = t.ConclaveID
		}
		if containerID != "" {
			result[containerID] = append(result[containerID], t)
		}
	}
	return result
}

// groupShipmentsByContainer groups shipments by their container ID (conclave or shipyard).
func (s *SummaryServiceImpl) groupShipmentsByContainer(shipments []*primary.Shipment) map[string][]*primary.Shipment {
	result := make(map[string][]*primary.Shipment)
	for _, ship := range shipments {
		if ship.ContainerID != "" {
			result[ship.ContainerID] = append(result[ship.ContainerID], ship)
		}
	}
	return result
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

// buildLibrarySummaryWithDebug counts tomes in the Library with debug tracking.
func (s *SummaryServiceImpl) buildLibrarySummaryWithDebug(ctx context.Context, tomes []*primary.Tome, expand bool, focusID string, addDebug func(string)) primary.LibrarySummary {
	var libTomes []primary.TomeSummary
	count := 0
	for _, t := range tomes {
		if t.ContainerType == "library" {
			if t.Status == "closed" {
				addDebug(fmt.Sprintf("Hidden: %s (%s) - status is closed", t.ID, t.Title))
				continue
			}
			count++
			if expand {
				// Expand notes if the tome itself is focused
				expandNotes := t.ID == focusID
				tomeSummary, err := s.buildTomeSummary(ctx, t, focusID, expandNotes)
				if err == nil {
					libTomes = append(libTomes, *tomeSummary)
				}
			}
		}
	}
	return primary.LibrarySummary{TomeCount: count, Tomes: libTomes}
}

// buildShipyardSummaryWithDebug counts shipments in the Shipyard with debug tracking.
func (s *SummaryServiceImpl) buildShipyardSummaryWithDebug(ctx context.Context, shipments []*primary.Shipment, expand bool, focusID string, addDebug func(string)) primary.ShipyardSummary {
	var yardShipments []primary.ShipmentSummary
	count := 0
	for _, ship := range shipments {
		if ship.ContainerType == "shipyard" {
			if ship.Status == "complete" {
				addDebug(fmt.Sprintf("Hidden: %s (%s) - status is complete", ship.ID, ship.Title))
				continue
			}
			count++
			if expand {
				shipSummary, err := s.buildShipmentSummary(ctx, ship, focusID)
				if err == nil {
					yardShipments = append(yardShipments, *shipSummary)
				}
			}
		}
	}
	return primary.ShipyardSummary{ShipmentCount: count, Shipments: yardShipments}
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
