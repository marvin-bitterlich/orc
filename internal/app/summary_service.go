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
) *SummaryServiceImpl {
	return &SummaryServiceImpl{
		commissionService: commissionService,
		tomeService:       tomeService,
		shipmentService:   shipmentService,
		taskService:       taskService,
		noteService:       noteService,
		workbenchService:  workbenchService,
		planService:       planService,
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
		if ship.Status == "closed" {
			addDebug(fmt.Sprintf("Hidden: %s (%s) - status is closed", ship.ID, ship.Title))
			continue
		}

		shipSummary, err := s.buildShipmentSummary(ctx, ship, req.FocusID)
		if err != nil {
			continue // Skip on error
		}
		shipmentSummaries = append(shipmentSummaries, *shipSummary)
	}

	// Determine if this is the focused commission (needed before tome expansion)
	isFocusedCommission := false
	focusIsCommission := req.FocusID == commission.ID
	focusIsShipment := false
	if req.FocusID != "" {
		if focusIsCommission {
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
				focusIsShipment = true
				break
			}
		}
	}

	// Build flat tome list
	var tomeSummaries []primary.TomeSummary
	for _, tome := range allTomes {
		if tome.Status == "closed" {
			addDebug(fmt.Sprintf("Hidden: %s (%s) - status is closed", tome.ID, tome.Title))
			continue
		}
		expandNotes := tome.ID == req.FocusID || focusIsCommission || focusIsShipment
		tomeSummary, err := s.buildTomeSummary(ctx, tome, req.FocusID, expandNotes)
		if err != nil {
			continue // Skip on error
		}
		tomeSummaries = append(tomeSummaries, *tomeSummary)
	}

	// Fetch commission-level notes (notes with no container)
	var noteSummaries []primary.NoteSummary
	commissionNotes, err := s.noteService.GetNotesByContainer(ctx, "commission", req.CommissionID)
	if err == nil {
		for _, note := range commissionNotes {
			if note.Status == "closed" {
				continue // Skip closed notes
			}
			noteSummaries = append(noteSummaries, primary.NoteSummary{
				ID:     note.ID,
				Title:  note.Title,
				Type:   note.Type,
				Status: note.Status,
				Pinned: note.Pinned,
			})
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
		Notes:               noteSummaries,
		DebugInfo:           debugInfo,
	}, nil
}

// buildTomeSummary creates a TomeSummary with note count.
// When expandNotes is true, includes the full Notes slice (for focused tomes).
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
						ID:     n.ID,
						Title:  n.Title,
						Type:   n.Type,
						Status: n.Status,
						Pinned: n.Pinned,
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
			if t.Status == "closed" {
				tasksDone++
			}
			// Include non-closed tasks for focused shipment
			if isFocused && t.Status != "closed" {
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

	// Get notes for this shipment (count open notes, expand if focused)
	noteCount := 0
	var noteSummaries []primary.NoteSummary
	notes, err := s.noteService.GetNotesByContainer(ctx, "shipment", ship.ID)
	if err == nil {
		for _, n := range notes {
			if n.Status != "closed" {
				noteCount++
				if isFocused {
					noteSummaries = append(noteSummaries, primary.NoteSummary{
						ID:     n.ID,
						Title:  n.Title,
						Type:   n.Type,
						Status: n.Status,
						Pinned: n.Pinned,
					})
				}
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
		NoteCount:  noteCount,
		Tasks:      taskSummaries,
		Notes:      noteSummaries,
	}, nil
}

// fetchTaskChildren populates the Plans for a task.
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
}

// Ensure SummaryServiceImpl implements the interface
var _ primary.SummaryService = (*SummaryServiceImpl)(nil)
