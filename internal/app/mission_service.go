package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	coremission "github.com/example/orc/internal/core/mission"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// MissionServiceImpl implements the MissionService interface.
type MissionServiceImpl struct {
	missionRepo   secondary.MissionRepository
	groveRepo     secondary.GroveRepository
	agentProvider secondary.AgentIdentityProvider
	executor      EffectExecutor
}

// NewMissionService creates a new MissionService with injected dependencies.
func NewMissionService(
	missionRepo secondary.MissionRepository,
	groveRepo secondary.GroveRepository,
	agentProvider secondary.AgentIdentityProvider,
	executor EffectExecutor,
) *MissionServiceImpl {
	return &MissionServiceImpl{
		missionRepo:   missionRepo,
		groveRepo:     groveRepo,
		agentProvider: agentProvider,
		executor:      executor,
	}
}

// CreateMission creates a new mission.
func (s *MissionServiceImpl) CreateMission(ctx context.Context, req primary.CreateMissionRequest) (*primary.CreateMissionResponse, error) {
	// 1. Get agent identity for guard
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	// 2. Check guard
	guardCtx := coremission.GuardContext{
		AgentType: coremission.AgentType(identity.Type),
		AgentID:   identity.FullID,
		MissionID: identity.MissionID,
	}
	if result := coremission.CanCreateMission(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 3. Generate ID using core business rule
	nextID, err := s.missionRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mission ID: %w", err)
	}

	// 4. Create mission record with pre-populated ID and initial status from core
	record := &secondary.MissionRecord{
		ID:          nextID,
		Title:       req.Title,
		Description: req.Description,
		Status:      string(coremission.InitialStatus()),
	}

	if err := s.missionRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create mission: %w", err)
	}

	// 5. Return response
	return &primary.CreateMissionResponse{
		MissionID: record.ID,
		Mission:   s.recordToMission(record),
	}, nil
}

// StartMission starts a mission with TMux session.
func (s *MissionServiceImpl) StartMission(ctx context.Context, req primary.StartMissionRequest) (*primary.StartMissionResponse, error) {
	// 1. Guard check
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	guardCtx := coremission.GuardContext{
		AgentType: coremission.AgentType(identity.Type),
		AgentID:   identity.FullID,
		MissionID: identity.MissionID,
	}
	if result := coremission.CanStartMission(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 2. Fetch mission
	mission, err := s.missionRepo.GetByID(ctx, req.MissionID)
	if err != nil {
		return nil, fmt.Errorf("mission not found: %w", err)
	}

	// 3. Fetch groves for mission
	groveRecords, err := s.groveRepo.GetByMission(ctx, req.MissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groves: %w", err)
	}

	// 4. Build planner input
	workspacePath := s.defaultWorkspacePath(req.MissionID)
	groves := make([]coremission.GrovePlanInput, len(groveRecords))
	for i, g := range groveRecords {
		groves[i] = coremission.GrovePlanInput{
			ID:          g.ID,
			Name:        g.Name,
			CurrentPath: g.WorktreePath,
			PathExists:  s.pathExists(g.WorktreePath),
		}
	}

	planInput := coremission.StartPlanInput{
		MissionID:     req.MissionID,
		WorkspacePath: workspacePath,
		Groves:        groves,
	}

	// 5. Generate plan (pure function)
	plan := coremission.GenerateStartPlan(planInput)

	// 6. Execute effects
	if err := s.executor.Execute(ctx, plan.Effects()); err != nil {
		return nil, fmt.Errorf("failed to execute start plan: %w", err)
	}

	return &primary.StartMissionResponse{
		Mission: s.recordToMission(mission),
	}, nil
}

// LaunchMission creates and starts mission infrastructure.
func (s *MissionServiceImpl) LaunchMission(ctx context.Context, req primary.LaunchMissionRequest) (*primary.LaunchMissionResponse, error) {
	// 1. Guard check
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	guardCtx := coremission.GuardContext{
		AgentType: coremission.AgentType(identity.Type),
		AgentID:   identity.FullID,
		MissionID: identity.MissionID,
	}
	if result := coremission.CanLaunchMission(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 2. Create mission first
	createResp, err := s.CreateMission(ctx, primary.CreateMissionRequest(req))
	if err != nil {
		return nil, err
	}

	// 3. Start mission
	_, err = s.StartMission(ctx, primary.StartMissionRequest{
		MissionID: createResp.MissionID,
	})
	if err != nil {
		return nil, err
	}

	return &primary.LaunchMissionResponse{
		MissionID: createResp.MissionID,
		Mission:   createResp.Mission,
	}, nil
}

// GetMission retrieves a mission by ID.
func (s *MissionServiceImpl) GetMission(ctx context.Context, missionID string) (*primary.Mission, error) {
	record, err := s.missionRepo.GetByID(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("mission not found: %w", err)
	}
	return s.recordToMission(record), nil
}

// ListMissions lists missions with optional filters.
func (s *MissionServiceImpl) ListMissions(ctx context.Context, filters primary.MissionFilters) ([]*primary.Mission, error) {
	records, err := s.missionRepo.List(ctx, secondary.MissionFilters{
		Status: filters.Status,
		Limit:  filters.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list missions: %w", err)
	}

	missions := make([]*primary.Mission, len(records))
	for i, r := range records {
		missions[i] = s.recordToMission(r)
	}
	return missions, nil
}

// CompleteMission marks a mission as complete.
func (s *MissionServiceImpl) CompleteMission(ctx context.Context, missionID string) error {
	// 1. Fetch mission to check state
	record, err := s.missionRepo.GetByID(ctx, missionID)
	if err != nil {
		return fmt.Errorf("mission not found: %w", err)
	}

	// 2. Guard check
	stateCtx := coremission.MissionStateContext{
		MissionID: missionID,
		IsPinned:  record.Pinned,
	}
	if result := coremission.CanCompleteMission(stateCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Apply status transition using core business logic
	transition := coremission.ApplyStatusTransition(coremission.StatusComplete, time.Now())
	record.Status = string(transition.NewStatus)
	if transition.CompletedAt != nil {
		record.CompletedAt = transition.CompletedAt.Format(time.RFC3339)
	}

	return s.missionRepo.Update(ctx, record)
}

// ArchiveMission archives a completed mission.
func (s *MissionServiceImpl) ArchiveMission(ctx context.Context, missionID string) error {
	// 1. Fetch mission to check state
	record, err := s.missionRepo.GetByID(ctx, missionID)
	if err != nil {
		return fmt.Errorf("mission not found: %w", err)
	}

	// 2. Guard check
	stateCtx := coremission.MissionStateContext{
		MissionID: missionID,
		IsPinned:  record.Pinned,
	}
	if result := coremission.CanArchiveMission(stateCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Apply status transition using core business logic
	transition := coremission.ApplyStatusTransition(coremission.StatusArchived, time.Now())
	record.Status = string(transition.NewStatus)

	return s.missionRepo.Update(ctx, record)
}

// UpdateMission updates mission title and/or description.
func (s *MissionServiceImpl) UpdateMission(ctx context.Context, req primary.UpdateMissionRequest) error {
	record, err := s.missionRepo.GetByID(ctx, req.MissionID)
	if err != nil {
		return fmt.Errorf("mission not found: %w", err)
	}

	if req.Title != "" {
		record.Title = req.Title
	}
	if req.Description != "" {
		record.Description = req.Description
	}

	return s.missionRepo.Update(ctx, record)
}

// DeleteMission deletes a mission.
func (s *MissionServiceImpl) DeleteMission(ctx context.Context, req primary.DeleteMissionRequest) error {
	// 1. Count dependents
	shipmentCount, err := s.missionRepo.CountShipments(ctx, req.MissionID)
	if err != nil {
		return fmt.Errorf("failed to count shipments: %w", err)
	}

	groves, err := s.groveRepo.GetByMission(ctx, req.MissionID)
	if err != nil {
		return fmt.Errorf("failed to get groves: %w", err)
	}

	// 2. Guard check
	deleteCtx := coremission.DeleteContext{
		MissionID:     req.MissionID,
		ShipmentCount: shipmentCount,
		GroveCount:    len(groves),
		ForceDelete:   req.Force,
	}
	if result := coremission.CanDeleteMission(deleteCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Delete
	return s.missionRepo.Delete(ctx, req.MissionID)
}

// PinMission pins a mission.
func (s *MissionServiceImpl) PinMission(ctx context.Context, missionID string) error {
	// 1. Check if mission exists
	record, err := s.missionRepo.GetByID(ctx, missionID)
	missionExists := err == nil

	// 2. Guard check
	pinCtx := coremission.PinContext{
		MissionID:     missionID,
		MissionExists: missionExists,
		IsPinned:      missionExists && record.Pinned,
	}
	if result := coremission.CanPinMission(pinCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Pin the mission
	return s.missionRepo.Pin(ctx, missionID)
}

// UnpinMission unpins a mission.
func (s *MissionServiceImpl) UnpinMission(ctx context.Context, missionID string) error {
	// 1. Check if mission exists
	record, err := s.missionRepo.GetByID(ctx, missionID)
	missionExists := err == nil

	// 2. Guard check
	pinCtx := coremission.PinContext{
		MissionID:     missionID,
		MissionExists: missionExists,
		IsPinned:      missionExists && record.Pinned,
	}
	if result := coremission.CanUnpinMission(pinCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Unpin the mission
	return s.missionRepo.Unpin(ctx, missionID)
}

// Helper methods

func (s *MissionServiceImpl) recordToMission(r *secondary.MissionRecord) *primary.Mission {
	return &primary.Mission{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		StartedAt:   r.StartedAt,
		CompletedAt: r.CompletedAt,
	}
}

func (s *MissionServiceImpl) defaultWorkspacePath(missionID string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "src", "missions", missionID)
}

func (s *MissionServiceImpl) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Ensure MissionServiceImpl implements the interface
var _ primary.MissionService = (*MissionServiceImpl)(nil)
