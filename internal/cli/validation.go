package cli

import (
	"fmt"
	"regexp"
	"strings"
)

// entityPrefixes maps entity types to their expected ID prefixes
var entityPrefixes = map[string]string{
	"tome":       "TOME",
	"shipment":   "SHIP",
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

// validateClaudeWorkspaceTrust removed - was only used by deprecated `orc commission start` command
