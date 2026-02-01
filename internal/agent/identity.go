package agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/orc/internal/config"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeGoblin   AgentType = "GOBLIN"
	AgentTypeIMP      AgentType = "IMP"
	AgentTypeWatchdog AgentType = "WATCHDOG"
)

// AgentIdentity represents a parsed agent ID
type AgentIdentity struct {
	Type   AgentType
	ID     string // "GOBLIN" for orchestrator, Workbench ID for IMP
	FullID string // Complete ID like "GOBLIN" or "IMP-BENCH-001"
}

// GetCurrentAgentID detects the current agent identity based on working directory context
// Simple logic: If place_id is BENCH-XXX → IMP, if GATE-XXX → Goblin, otherwise → Goblin
func GetCurrentAgentID() (*AgentIdentity, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check for config in current directory
	cfg, err := config.LoadConfig(cwd)
	if err == nil && cfg.PlaceID != "" {
		placeType := config.GetPlaceType(cfg.PlaceID)
		switch placeType {
		case config.PlaceTypeWorkbench:
			// We're in a workbench - this is an IMP
			return &AgentIdentity{
				Type:   AgentTypeIMP,
				ID:     cfg.PlaceID,
				FullID: fmt.Sprintf("IMP-%s", cfg.PlaceID),
			}, nil
		case config.PlaceTypeGatehouse:
			// We're in a gatehouse - this is a Goblin
			return &AgentIdentity{
				Type:   AgentTypeGoblin,
				ID:     cfg.PlaceID,
				FullID: fmt.Sprintf("GOBLIN-%s", cfg.PlaceID),
			}, nil
		case config.PlaceTypeKennel:
			// We're in a kennel - this is a Watchdog
			return &AgentIdentity{
				Type:   AgentTypeWatchdog,
				ID:     cfg.PlaceID,
				FullID: fmt.Sprintf("WATCHDOG-%s", cfg.PlaceID),
			}, nil
		}
	}

	// Not in a recognized place - we're a Goblin (orchestrator) by default
	// Goblin can work anywhere: commission workspaces, ORC repo, anywhere
	return &AgentIdentity{
		Type:   AgentTypeGoblin,
		ID:     "GOBLIN",
		FullID: "GOBLIN",
	}, nil
}

// ParseAgentID parses an agent ID string like "GOBLIN" or "IMP-BENCH-001"
func ParseAgentID(agentID string) (*AgentIdentity, error) {
	// Special case: GOBLIN or ORC (backwards compat) has no parts
	if agentID == "GOBLIN" || agentID == "ORC" {
		return &AgentIdentity{
			Type:   AgentTypeGoblin,
			ID:     "GOBLIN",
			FullID: "GOBLIN",
		}, nil
	}

	parts := strings.SplitN(agentID, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid agent ID format: %s (expected GOBLIN or IMP-WORKBENCH-ID)", agentID)
	}

	agentType := AgentType(parts[0])
	id := parts[1]

	switch agentType {
	case AgentTypeIMP:
		// For IMP, workbench IDs are like BENCH-001
		// Commission context is resolved via DB lookup when needed
		return &AgentIdentity{
			Type:   AgentTypeIMP,
			ID:     id,
			FullID: agentID,
		}, nil
	case AgentTypeGoblin:
		// For GOBLIN, gatehouse IDs are like GATE-003
		return &AgentIdentity{
			Type:   AgentTypeGoblin,
			ID:     id,
			FullID: agentID,
		}, nil
	case AgentTypeWatchdog:
		// For WATCHDOG, kennel IDs are like KENNEL-014
		return &AgentIdentity{
			Type:   AgentTypeWatchdog,
			ID:     id,
			FullID: agentID,
		}, nil
	default:
		return nil, fmt.Errorf("unknown agent type: %s (expected GOBLIN, IMP, or WATCHDOG)", agentType)
	}
}
