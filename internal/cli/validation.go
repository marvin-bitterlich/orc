package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// entityPrefixes maps entity types to their expected ID prefixes
var entityPrefixes = map[string]string{
	"tome":       "TOME",
	"shipment":   "SHIP",
	"conclave":   "CON",
	"commission": "COMM",
	"task":       "TASK",
	"note":       "NOTE",
	"workbench":  "BENCH",
	"workshop":   "WORK",
}

// validateEntityID checks if an ID has the correct prefix format.
// Returns an error with helpful message if the ID appears to be a short ID.
func validateEntityID(id, entityType string) error {
	if id == "" {
		return nil // Empty is OK, let other validation handle required fields
	}

	prefix, ok := entityPrefixes[entityType]
	if !ok {
		return nil // Unknown entity type, skip validation
	}

	expectedPattern := prefix + "-"
	if strings.HasPrefix(id, expectedPattern) {
		return nil // Valid format
	}

	// Check if it looks like a short ID (just digits)
	if matched, _ := regexp.MatchString(`^\d+$`, id); matched {
		return fmt.Errorf("invalid %s ID '%s'. Use full ID format: %s-%s", entityType, id, prefix, id)
	}

	// Check if it's using wrong case
	if strings.HasPrefix(strings.ToUpper(id), expectedPattern) {
		return fmt.Errorf("invalid %s ID '%s'. IDs are case-sensitive, use: %s", entityType, id, strings.ToUpper(id))
	}

	// Generic invalid format
	return fmt.Errorf("invalid %s ID '%s'. Expected format: %s-xxx", entityType, id, prefix)
}

// validateClaudeWorkspaceTrust checks if Claude Code settings include required directories
// for ORC workbenches and commissions. Returns nil if valid, error with fix instructions if not.
func validateClaudeWorkspaceTrust() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Check if settings file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return fmt.Errorf(`~/.claude/settings.json not found

Claude Code workspace trust is required for ORC to function.

Create the file with:

  cat > ~/.claude/settings.json <<'EOF'
  {
    "permissions": {
      "additionalDirectories": [
        "~/src/worktrees",
        "~/src/factories"
      ]
    }
  }
  EOF

See INSTALL.md for detailed setup instructions.`)
	}

	// Read and parse settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read ~/.claude/settings.json: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("~/.claude/settings.json is not valid JSON: %w\n\nFix the JSON syntax and try again", err)
	}

	// Check for permissions.additionalDirectories
	permissions, ok := settings["permissions"].(map[string]interface{})
	if !ok {
		return fmt.Errorf(`permissions not configured in ~/.claude/settings.json

Required configuration:

  {
    "permissions": {
      "additionalDirectories": [
        "~/src/worktrees",
        "~/src/factories"
      ]
    }
  }

Add this to your settings.json or run 'orc doctor' for help.`)
	}

	additionalDirs, ok := permissions["additionalDirectories"].([]interface{})
	if !ok {
		return fmt.Errorf(`permissions.additionalDirectories not configured

ORC requires Claude to trust specific directories.

Add to ~/.claude/settings.json:

  "permissions": {
    "additionalDirectories": [
      "~/src/worktrees",
      "~/src/factories"
    ]
  }`)
	}

	// Check for required directories
	requiredDirs := []string{"~/src/worktrees", "~/src/factories"}
	foundDirs := make(map[string]bool)

	for _, dir := range additionalDirs {
		if dirStr, ok := dir.(string); ok {
			foundDirs[dirStr] = true
		}
	}

	var missingDirs []string
	for _, required := range requiredDirs {
		if !foundDirs[required] {
			missingDirs = append(missingDirs, required)
		}
	}

	if len(missingDirs) > 0 {
		return fmt.Errorf(`Missing trusted directories in ~/.claude/settings.json:
  %s

These directories are required for ORC workbenches and commissions.

Add them to permissions.additionalDirectories:

  jq '.permissions.additionalDirectories += ["%s"]' \
    ~/.claude/settings.json > ~/.claude/settings.json.tmp && \
    mv ~/.claude/settings.json.tmp ~/.claude/settings.json`,
			strings.Join(missingDirs, "\n  "),
			strings.Join(missingDirs, "\", \""))
	}

	return nil
}
