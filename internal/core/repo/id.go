// Package repo contains the pure business logic for repository operations.
// This is part of the Functional Core - no I/O, only pure functions.
package repo

import "fmt"

// GenerateRepoID generates a repository ID from the current max number.
// This is a pure function that defines the ID format as a business rule.
// The format is REPO-XXX where XXX is a zero-padded 3-digit number.
func GenerateRepoID(currentMax int) string {
	return fmt.Sprintf("REPO-%03d", currentMax+1)
}

// ParseRepoNumber extracts the numeric portion from a repository ID.
// Returns -1 if the ID format is invalid.
func ParseRepoNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "REPO-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
