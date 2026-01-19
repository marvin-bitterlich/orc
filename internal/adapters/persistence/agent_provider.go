package persistence

import (
	"context"

	"github.com/example/orc/internal/agent"
	"github.com/example/orc/internal/ports/secondary"
)

// AgentIdentityProviderAdapter wraps the agent package to implement AgentIdentityProvider.
type AgentIdentityProviderAdapter struct{}

// NewAgentIdentityProvider creates a new AgentIdentityProviderAdapter.
func NewAgentIdentityProvider() *AgentIdentityProviderAdapter {
	return &AgentIdentityProviderAdapter{}
}

// GetCurrentIdentity returns the identity of the current agent.
func (p *AgentIdentityProviderAdapter) GetCurrentIdentity(ctx context.Context) (*secondary.AgentIdentity, error) {
	identity, err := agent.GetCurrentAgentID()
	if err != nil {
		return nil, err
	}

	return &secondary.AgentIdentity{
		Type:      secondary.AgentType(identity.Type),
		ID:        identity.ID,
		FullID:    identity.FullID,
		MissionID: identity.MissionID,
	}, nil
}

// Ensure AgentIdentityProviderAdapter implements the interface
var _ secondary.AgentIdentityProvider = (*AgentIdentityProviderAdapter)(nil)
