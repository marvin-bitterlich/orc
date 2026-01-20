// Package mission contains the pure business logic for mission operations.
// This is part of the Functional Core - no I/O, only pure functions.
package mission

import "fmt"

// GenerateMissionID generates a mission ID from the current max number.
// This is a pure function that defines the ID format as a business rule.
// The format is MISSION-XXX where XXX is a zero-padded 3-digit number.
func GenerateMissionID(currentMax int) string {
	return fmt.Sprintf("MISSION-%03d", currentMax+1)
}

// ParseMissionNumber extracts the numeric portion from a mission ID.
// Returns -1 if the ID format is invalid.
func ParseMissionNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "MISSION-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
