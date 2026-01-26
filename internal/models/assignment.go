package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ShipmentAssignment represents a shipment (with tasks) assigned to a workbench
type ShipmentAssignment struct {
	ShipmentID          string       `json:"shipment_id"`
	ShipmentTitle       string       `json:"shipment_title"`
	ShipmentDescription string       `json:"shipment_description"`
	CommissionID        string       `json:"commission_id"`
	AssignedBy          string       `json:"assigned_by"`
	AssignedAt          string       `json:"assigned_at"`
	Status              string       `json:"status"` // assigned, in_progress, complete
	Tasks               []TaskInfo   `json:"tasks"`
	Progress            ProgressInfo `json:"progress"`
}

// TaskInfo represents a task within a shipment
type TaskInfo struct {
	TaskID      string `json:"task_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Type        string `json:"type,omitempty"`
	ClaimedAt   string `json:"claimed_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// ProgressInfo tracks progress across shipment's tasks
type ProgressInfo struct {
	TotalTasks      int `json:"total_tasks"`
	CompletedTasks  int `json:"completed_tasks"`
	InProgressTasks int `json:"in_progress_tasks"`
	ReadyTasks      int `json:"ready_tasks"`
}

// ReadShipmentAssignment reads a shipment assignment from a workbench's .orc directory
func ReadShipmentAssignment(workbenchDir string) (*ShipmentAssignment, error) {
	path := filepath.Join(workbenchDir, ".orc", "assigned-work.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read assignment file: %w", err)
	}

	var assignment ShipmentAssignment
	if err := json.Unmarshal(data, &assignment); err != nil {
		return nil, fmt.Errorf("failed to parse shipment assignment: %w", err)
	}

	return &assignment, nil
}

// UpdateShipmentAssignmentStatus updates the status of a shipment assignment file
func UpdateShipmentAssignmentStatus(workbenchDir, status string) error {
	assignment, err := ReadShipmentAssignment(workbenchDir)
	if err != nil {
		return err
	}

	assignment.Status = status

	path := filepath.Join(workbenchDir, ".orc", "assigned-work.json")
	data, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal shipment assignment: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write assignment file: %w", err)
	}

	return nil
}

// WriteShipmentAssignment writes a shipment with tasks to assignment file
func WriteShipmentAssignment(workbenchDir string, shipment *Shipment, tasks []*Task, assignedBy string) error {
	// Build task info
	var taskInfos []TaskInfo
	completedTasks := 0
	inProgressTasks := 0
	readyTasks := 0

	for _, task := range tasks {
		taskInfo := TaskInfo{
			TaskID: task.ID,
			Title:  task.Title,
			Status: task.Status,
		}
		if task.Type.Valid {
			taskInfo.Type = task.Type.String
		}
		if task.ClaimedAt.Valid {
			taskInfo.ClaimedAt = task.ClaimedAt.Time.Format(time.RFC3339)
		}
		if task.CompletedAt.Valid {
			taskInfo.CompletedAt = task.CompletedAt.Time.Format(time.RFC3339)
		}

		taskInfos = append(taskInfos, taskInfo)

		// Update progress counts
		switch task.Status {
		case "complete":
			completedTasks++
		case "implement", "in_progress":
			inProgressTasks++
		case "ready":
			readyTasks++
		}
	}

	description := ""
	if shipment.Description.Valid {
		description = shipment.Description.String
	}

	assignment := &ShipmentAssignment{
		ShipmentID:          shipment.ID,
		ShipmentTitle:       shipment.Title,
		ShipmentDescription: description,
		CommissionID:        shipment.CommissionID,
		AssignedBy:          assignedBy,
		AssignedAt:          time.Now().Format(time.RFC3339),
		Status:              "assigned",
		Tasks:               taskInfos,
		Progress: ProgressInfo{
			TotalTasks:      len(tasks),
			CompletedTasks:  completedTasks,
			InProgressTasks: inProgressTasks,
			ReadyTasks:      readyTasks,
		},
	}

	orcDir := filepath.Join(workbenchDir, ".orc")
	if err := os.MkdirAll(orcDir, 0755); err != nil {
		return fmt.Errorf("failed to create .orc directory: %w", err)
	}

	path := filepath.Join(orcDir, "assigned-work.json")
	data, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal shipment assignment: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write assignment file: %w", err)
	}

	return nil
}
