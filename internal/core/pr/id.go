// Package pr contains the pure business logic for pull request operations.
// This is part of the Functional Core - no I/O, only pure functions.
package pr

import "fmt"

// GeneratePRID generates a pull request ID from the current max number.
// This is a pure function that defines the ID format as a business rule.
// The format is PR-XXX where XXX is a zero-padded 3-digit number.
func GeneratePRID(currentMax int) string {
	return fmt.Sprintf("PR-%03d", currentMax+1)
}

// ParsePRNumber extracts the numeric portion from a pull request ID.
// Returns -1 if the ID format is invalid.
func ParsePRNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "PR-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
