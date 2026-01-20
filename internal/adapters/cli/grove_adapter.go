package cli

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/example/orc/internal/ports/primary"
)

// GroveAdapter is a thin adapter that translates CLI operations to GroveService calls.
// It depends only on the GroveService interface, enabling easy testing with mocks.
type GroveAdapter struct {
	service primary.GroveService
	out     io.Writer
}

// NewGroveAdapter creates a new GroveAdapter with the given service.
func NewGroveAdapter(service primary.GroveService, out io.Writer) *GroveAdapter {
	return &GroveAdapter{
		service: service,
		out:     out,
	}
}

// List lists groves with optional mission filter.
// When missionID is empty, all groves are returned.
func (a *GroveAdapter) List(ctx context.Context, missionID string) ([]*primary.Grove, error) {
	groves, err := a.service.ListGroves(ctx, primary.GroveFilters{
		MissionID: missionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list groves: %w", err)
	}

	if len(groves) == 0 {
		fmt.Fprintln(a.out, "No groves found.")
		fmt.Fprintln(a.out)
		fmt.Fprintln(a.out, "Create your first grove:")
		fmt.Fprintln(a.out, "  orc grove create my-grove --repos main-app --mission MISSION-001")
		return groves, nil
	}

	w := tabwriter.NewWriter(a.out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tMISSION\tSTATUS\tPATH")
	fmt.Fprintln(w, "--\t----\t-------\t------\t----")

	for _, grove := range groves {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			grove.ID,
			grove.Name,
			grove.MissionID,
			grove.Status,
			grove.Path,
		)
	}

	w.Flush()
	return groves, nil
}

// Show displays details for a single grove.
func (a *GroveAdapter) Show(ctx context.Context, groveID string) (*primary.Grove, error) {
	grove, err := a.service.GetGrove(ctx, groveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get grove: %w", err)
	}

	fmt.Fprintf(a.out, "\nGrove: %s\n", grove.ID)
	fmt.Fprintf(a.out, "Name:    %s\n", grove.Name)
	fmt.Fprintf(a.out, "Mission: %s\n", grove.MissionID)
	fmt.Fprintf(a.out, "Path:    %s\n", grove.Path)
	fmt.Fprintf(a.out, "Status:  %s\n", grove.Status)
	fmt.Fprintf(a.out, "Created: %s\n", grove.CreatedAt)
	fmt.Fprintln(a.out)

	return grove, nil
}

// Rename renames a grove in the database.
// Note: Config file update is handled by the CLI layer.
func (a *GroveAdapter) Rename(ctx context.Context, groveID, newName string) (oldName string, err error) {
	// Get grove before rename (for old name display)
	grove, err := a.service.GetGrove(ctx, groveID)
	if err != nil {
		return "", fmt.Errorf("failed to get grove: %w", err)
	}

	oldName = grove.Name

	err = a.service.RenameGrove(ctx, primary.RenameGroveRequest{
		GroveID: groveID,
		NewName: newName,
	})
	if err != nil {
		return "", err
	}

	fmt.Fprintf(a.out, "✓ Grove %s renamed\n", groveID)
	fmt.Fprintf(a.out, "  %s → %s\n", oldName, newName)

	return oldName, nil
}

// Delete deletes a grove from the database.
// Note: Worktree removal is handled by the CLI layer.
func (a *GroveAdapter) Delete(ctx context.Context, groveID string, force bool) (*primary.Grove, error) {
	// Get grove details before deleting (for output and worktree path)
	grove, err := a.service.GetGrove(ctx, groveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get grove: %w", err)
	}

	err = a.service.DeleteGrove(ctx, primary.DeleteGroveRequest{
		GroveID: groveID,
		Force:   force,
	})
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(a.out, "✓ Deleted grove %s: %s\n", grove.ID, grove.Name)

	return grove, nil
}

// GetGrove retrieves a grove by ID (for use by CLI when it needs grove data).
func (a *GroveAdapter) GetGrove(ctx context.Context, groveID string) (*primary.Grove, error) {
	return a.service.GetGrove(ctx, groveID)
}
