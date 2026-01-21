package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/wire"
)

// FocusCmd returns the focus command
func FocusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus [container-id]",
		Short: "Set or show the currently focused container",
		Long: `Focus on a specific container (Shipment, Conclave, Investigation, or Tome).

The focused container appears in 'orc prime' output and can be used as default
for other commands.

Container types are auto-detected from ID prefix:
  SHIP-*  → Shipment (execution work)
  CON-*   → Conclave (ideation session)
  INV-*   → Investigation (research)
  TOME-*  → Tome (knowledge collection)

Examples:
  orc focus CON-001     # Focus on a conclave
  orc focus SHIP-178    # Focus on a shipment
  orc focus             # Clear focus
  orc focus --show      # Show current focus`,
		Args: cobra.MaximumNArgs(1),
		RunE: runFocus,
	}
	cmd.Flags().Bool("show", false, "Show current focus without changing it")
	return cmd
}

func runFocus(cmd *cobra.Command, args []string) error {
	showOnly, _ := cmd.Flags().GetBool("show")

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Try to load config from cwd first, then fall back to home dir
	cfg, configDir, err := loadConfigWithDir(cwd)
	if err != nil {
		// Try home directory for global config
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return fmt.Errorf("no config found and cannot access home directory")
		}
		cfg, configDir, err = loadConfigWithDir(homeDir)
		if err != nil {
			return fmt.Errorf("no ORC config found in current directory or home directory")
		}
	}

	if showOnly {
		return showCurrentFocus(cfg)
	}

	if len(args) == 0 {
		// Clear focus
		return clearFocus(cfg, configDir)
	}

	// Set focus
	containerID := args[0]
	containerType, title, err := validateAndGetInfo(containerID)
	if err != nil {
		return err
	}

	return setFocus(cfg, configDir, containerID, containerType, title)
}

// loadConfigWithDir loads config and returns both config and the directory it was loaded from
func loadConfigWithDir(dir string) (*config.Config, string, error) {
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return nil, "", err
	}
	return cfg, dir, nil
}

// validateAndGetInfo validates the container ID exists and returns its type and title
func validateAndGetInfo(id string) (containerType string, title string, err error) {
	ctx := context.Background()
	switch {
	case strings.HasPrefix(id, "SHIP-"):
		ship, err := wire.ShipmentService().GetShipment(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("shipment %s not found", id)
		}
		return "Shipment", ship.Title, nil

	case strings.HasPrefix(id, "CON-"):
		con, err := wire.ConclaveService().GetConclave(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("conclave %s not found", id)
		}
		return "Conclave", con.Title, nil

	case strings.HasPrefix(id, "INV-"):
		inv, err := wire.InvestigationService().GetInvestigation(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("investigation %s not found", id)
		}
		return "Investigation", inv.Title, nil

	case strings.HasPrefix(id, "TOME-"):
		tome, err := wire.TomeService().GetTome(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("tome %s not found", id)
		}
		return "Tome", tome.Title, nil

	default:
		return "", "", fmt.Errorf("unknown container type for ID: %s (expected SHIP-*, CON-*, INV-*, or TOME-*)", id)
	}
}

// showCurrentFocus displays the current focus
func showCurrentFocus(cfg *config.Config) error {
	focusID := getCurrentFocus(cfg)

	if focusID == "" {
		fmt.Println("No focus set")
		fmt.Println("\nSet focus with: orc focus <container-id>")
		return nil
	}

	containerType, title, err := validateAndGetInfo(focusID)
	if err != nil {
		// Focus is set but container no longer exists - graceful degradation, not an error
		fmt.Printf("Focus: %s (container not found - may have been deleted)\n", focusID)
		return nil //nolint:nilerr // intentional: show info even if container deleted
	}

	fmt.Printf("Focus: %s\n", focusID)
	fmt.Printf("  %s: %s\n", containerType, title)
	return nil
}

// getCurrentFocus gets the current focus from config based on config type
func getCurrentFocus(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	switch cfg.Type {
	case config.TypeMission:
		if cfg.Mission != nil {
			return cfg.Mission.CurrentFocus
		}
	case config.TypeGrove:
		if cfg.Grove != nil {
			return cfg.Grove.CurrentFocus
		}
	case config.TypeGlobal:
		if cfg.State != nil {
			return cfg.State.CurrentFocus
		}
	}
	return ""
}

// setFocus sets the focus in the config
func setFocus(cfg *config.Config, configDir, containerID, containerType, title string) error {
	switch cfg.Type {
	case config.TypeMission:
		if cfg.Mission != nil {
			cfg.Mission.CurrentFocus = containerID
		}
	case config.TypeGrove:
		if cfg.Grove != nil {
			cfg.Grove.CurrentFocus = containerID
		}
	case config.TypeGlobal:
		if cfg.State != nil {
			cfg.State.CurrentFocus = containerID
		}
	}

	if err := config.SaveConfig(configDir, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Focused on %s: %s\n", containerType, containerID)
	fmt.Printf("  %s\n", title)
	fmt.Println("\nRun 'orc prime' to see updated context.")
	return nil
}

// clearFocus clears the current focus
func clearFocus(cfg *config.Config, configDir string) error {
	switch cfg.Type {
	case config.TypeMission:
		if cfg.Mission != nil {
			cfg.Mission.CurrentFocus = ""
		}
	case config.TypeGrove:
		if cfg.Grove != nil {
			cfg.Grove.CurrentFocus = ""
		}
	case config.TypeGlobal:
		if cfg.State != nil {
			cfg.State.CurrentFocus = ""
		}
	}

	if err := config.SaveConfig(configDir, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Focus cleared")
	return nil
}

// GetCurrentFocus is exported for use by other commands (e.g., prime)
func GetCurrentFocus(cfg *config.Config) string {
	return getCurrentFocus(cfg)
}

// GetFocusInfo returns the type and title for a focus ID, or empty strings if invalid
func GetFocusInfo(focusID string) (containerType, title, status string) {
	if focusID == "" {
		return "", "", ""
	}

	ctx := context.Background()
	switch {
	case strings.HasPrefix(focusID, "SHIP-"):
		if ship, err := wire.ShipmentService().GetShipment(ctx, focusID); err == nil {
			return "Shipment", ship.Title, ship.Status
		}
	case strings.HasPrefix(focusID, "CON-"):
		if con, err := wire.ConclaveService().GetConclave(ctx, focusID); err == nil {
			return "Conclave", con.Title, con.Status
		}
	case strings.HasPrefix(focusID, "INV-"):
		if inv, err := wire.InvestigationService().GetInvestigation(ctx, focusID); err == nil {
			return "Investigation", inv.Title, inv.Status
		}
	case strings.HasPrefix(focusID, "TOME-"):
		if tome, err := wire.TomeService().GetTome(ctx, focusID); err == nil {
			return "Tome", tome.Title, tome.Status
		}
	}
	return "", "", ""
}
