// Package app contains the application services that orchestrate business logic.
package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// CommissionOrchestrationService handles complex commission infrastructure operations.
// It implements the plan/apply pattern for idempotent infrastructure management.
type CommissionOrchestrationService struct {
	commissionSvc primary.CommissionService
	agentProvider secondary.AgentIdentityProvider
}

// NewCommissionOrchestrationService creates a new orchestration service.
func NewCommissionOrchestrationService(commissionSvc primary.CommissionService, agentProvider secondary.AgentIdentityProvider) *CommissionOrchestrationService {
	return &CommissionOrchestrationService{
		commissionSvc: commissionSvc,
		agentProvider: agentProvider,
	}
}

// LoadCommissionState loads the commission from the database.
func (s *CommissionOrchestrationService) LoadCommissionState(ctx context.Context, commissionID string) (*primary.CommissionState, error) {
	commission, err := s.commissionSvc.GetCommission(ctx, commissionID)
	if err != nil {
		return nil, fmt.Errorf("commission not found: %w", err)
	}

	return &primary.CommissionState{
		Commission: commission,
	}, nil
}

// AnalyzeInfrastructure generates a plan for setting up commission infrastructure.
func (s *CommissionOrchestrationService) AnalyzeInfrastructure(state *primary.CommissionState, workspacePath string) *primary.InfrastructurePlan {
	return &primary.InfrastructurePlan{
		WorkspacePath:  workspacePath,
		WorkbenchesDir: "",
	}
}

// ApplyInfrastructure applies the infrastructure plan.
func (s *CommissionOrchestrationService) ApplyInfrastructure(ctx context.Context, plan *primary.InfrastructurePlan) *primary.InfrastructureApplyResult {
	return &primary.InfrastructureApplyResult{}
}

// PlanTmuxSession generates a plan for the TMux session.
func (s *CommissionOrchestrationService) PlanTmuxSession(state *primary.CommissionState, workspacePath, sessionName string, sessionExists bool, windowChecker primary.TmuxWindowChecker) *primary.TmuxSessionPlan {
	return &primary.TmuxSessionPlan{
		SessionName:   sessionName,
		WorkingDir:    workspacePath,
		SessionExists: sessionExists,
	}
}
