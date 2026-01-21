package agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/context"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeORC AgentType = "ORC"
	AgentTypeIMP AgentType = "IMP"
)

// AgentIdentity represents a parsed agent ID
type AgentIdentity struct {
	Type      AgentType
	ID        string // "ORC" for orchestrator, Grove ID for IMP
	FullID    string // Complete ID like "ORC" or "IMP-GROVE-001"
	MissionID string // Mission ID (empty for ORC outside mission, populated for IMP)
}

// GetCurrentAgentID detects the current agent identity based on working directory context
// Simple logic: If in grove → IMP, otherwise → ORC
func GetCurrentAgentID() (*AgentIdentity, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check for grove config first (most specific)
	cfg, err := config.LoadConfigWithFallback(cwd)
	if err == nil && cfg.Type == config.TypeGrove {
		// We're in a grove - this is an IMP
		return &AgentIdentity{
			Type:      AgentTypeIMP,
			ID:        cfg.Grove.GroveID,
			FullID:    fmt.Sprintf("IMP-%s", cfg.Grove.GroveID),
			MissionID: cfg.Grove.MissionID,
		}, nil
	}

	// Not in a grove - we're ORC (orchestrator)
	// ORC can work anywhere: mission workspaces, ORC repo, anywhere
	missionCtx, _ := context.DetectMissionContext()
	missionID := ""
	if missionCtx != nil {
		missionID = missionCtx.MissionID
	}

	return &AgentIdentity{
		Type:      AgentTypeORC,
		ID:        "ORC",
		FullID:    "ORC",
		MissionID: missionID, // Populated if in mission context, empty otherwise
	}, nil
}

// ParseAgentID parses an agent ID string like "ORC" or "IMP-GROVE-001"
func ParseAgentID(agentID string) (*AgentIdentity, error) {
	// Special case: ORC has no parts
	if agentID == "ORC" {
		return &AgentIdentity{
			Type:      AgentTypeORC,
			ID:        "ORC",
			FullID:    "ORC",
			MissionID: "", // ORC can work across missions
		}, nil
	}

	parts := strings.SplitN(agentID, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid agent ID format: %s (expected ORC or IMP-GROVE-ID)", agentID)
	}

	agentType := AgentType(parts[0])
	id := parts[1]

	switch agentType {
	case AgentTypeIMP:
		// For IMP, we need to extract mission ID from grove ID
		// Grove IDs are like GROVE-001, we need to look up the mission
		// For now, return partial identity (caller must resolve mission)
		return &AgentIdentity{
			Type:   AgentTypeIMP,
			ID:     id,
			FullID: agentID,
		}, nil
	default:
		return nil, fmt.Errorf("unknown agent type: %s (expected ORC or IMP)", agentType)
	}
}

// ResolveTMuxTarget converts an agent ID to a tmux target string
func ResolveTMuxTarget(agentID string, groveName string) (string, error) {
	identity, err := ParseAgentID(agentID)
	if err != nil {
		return "", err
	}

	if identity.Type == AgentTypeORC {
		// ORC always in ORC session, window 1, pane 1
		return "ORC:1.1", nil
	}

	// For IMP, need grove name and mission ID
	if identity.MissionID == "" || groveName == "" {
		return "", fmt.Errorf("IMP target requires mission ID and grove name")
	}

	// Window named by grove, pane 2 is Claude
	return fmt.Sprintf("orc-%s:%s.2", identity.MissionID, groveName), nil
}
