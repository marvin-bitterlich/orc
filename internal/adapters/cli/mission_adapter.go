// Package cli provides thin CLI adapters that translate between CLI concerns
// and application services. Adapters handle argument parsing, output formatting,
// but delegate business logic to services.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/example/orc/internal/ports/primary"
)

// MissionAdapter is a thin adapter that translates CLI operations to MissionService calls.
// It depends only on the MissionService interface, enabling easy testing with mocks.
type MissionAdapter struct {
	service primary.MissionService
	out     io.Writer
}

// NewMissionAdapter creates a new MissionAdapter with the given service.
func NewMissionAdapter(service primary.MissionService, out io.Writer) *MissionAdapter {
	return &MissionAdapter{
		service: service,
		out:     out,
	}
}

// Create creates a new mission.
func (a *MissionAdapter) Create(ctx context.Context, title, description string) error {
	resp, err := a.service.CreateMission(ctx, primary.CreateMissionRequest{
		Title:       title,
		Description: description,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "✓ Created mission %s: %s\n", resp.MissionID, resp.Mission.Title)
	return nil
}

// List lists missions with optional status filter.
func (a *MissionAdapter) List(ctx context.Context, status string) error {
	missions, err := a.service.ListMissions(ctx, primary.MissionFilters{
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to list missions: %w", err)
	}

	if len(missions) == 0 {
		fmt.Fprintln(a.out, "No missions found")
		return nil
	}

	fmt.Fprintf(a.out, "\n%-15s %-10s %s\n", "ID", "STATUS", "TITLE")
	fmt.Fprintln(a.out, "────────────────────────────────────────────────────────────────")
	for _, m := range missions {
		fmt.Fprintf(a.out, "%-15s %-10s %s\n", m.ID, m.Status, m.Title)
	}
	fmt.Fprintln(a.out)

	return nil
}

// Show displays details for a single mission.
// Note: Related entities (shipments, groves) are fetched separately by the CLI layer.
func (a *MissionAdapter) Show(ctx context.Context, missionID string) (*primary.Mission, error) {
	mission, err := a.service.GetMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mission: %w", err)
	}

	fmt.Fprintf(a.out, "\nMission: %s\n", mission.ID)
	fmt.Fprintf(a.out, "Title:   %s\n", mission.Title)
	fmt.Fprintf(a.out, "Status:  %s\n", mission.Status)
	if mission.Description != "" {
		fmt.Fprintf(a.out, "Description: %s\n", mission.Description)
	}
	fmt.Fprintf(a.out, "Created: %s\n", mission.CreatedAt)
	if mission.CompletedAt != "" {
		fmt.Fprintf(a.out, "Completed: %s\n", mission.CompletedAt)
	}
	fmt.Fprintln(a.out)

	return mission, nil
}

// Update updates a mission's title and/or description.
func (a *MissionAdapter) Update(ctx context.Context, missionID, title, description string) error {
	if title == "" && description == "" {
		return fmt.Errorf("must specify at least --title or --description")
	}

	err := a.service.UpdateMission(ctx, primary.UpdateMissionRequest{
		MissionID:   missionID,
		Title:       title,
		Description: description,
	})
	if err != nil {
		return fmt.Errorf("failed to update mission: %w", err)
	}

	fmt.Fprintf(a.out, "✓ Mission %s updated\n", missionID)
	return nil
}

// Complete marks a mission as complete.
func (a *MissionAdapter) Complete(ctx context.Context, missionID string) error {
	err := a.service.CompleteMission(ctx, missionID)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "✓ Mission %s marked as complete\n", missionID)
	return nil
}

// Archive archives a mission.
func (a *MissionAdapter) Archive(ctx context.Context, missionID string) error {
	err := a.service.ArchiveMission(ctx, missionID)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "✓ Mission %s archived\n", missionID)
	return nil
}

// Delete deletes a mission.
func (a *MissionAdapter) Delete(ctx context.Context, missionID string, force bool) error {
	// Get mission details before deleting (for output)
	mission, err := a.service.GetMission(ctx, missionID)
	if err != nil {
		return fmt.Errorf("failed to get mission: %w", err)
	}

	err = a.service.DeleteMission(ctx, primary.DeleteMissionRequest{
		MissionID: missionID,
		Force:     force,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "✓ Deleted mission %s: %s\n", mission.ID, mission.Title)
	return nil
}

// Pin pins a mission.
func (a *MissionAdapter) Pin(ctx context.Context, missionID string) error {
	err := a.service.PinMission(ctx, missionID)
	if err != nil {
		return fmt.Errorf("failed to pin mission: %w", err)
	}

	fmt.Fprintf(a.out, "✓ Mission %s pinned 📌\n", missionID)
	return nil
}

// Unpin unpins a mission.
func (a *MissionAdapter) Unpin(ctx context.Context, missionID string) error {
	err := a.service.UnpinMission(ctx, missionID)
	if err != nil {
		return fmt.Errorf("failed to unpin mission: %w", err)
	}

	fmt.Fprintf(a.out, "✓ Mission %s unpinned\n", missionID)
	return nil
}
