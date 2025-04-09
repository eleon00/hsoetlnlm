package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eleon00/hsoetlnlm/internal/data"
)

// WorkflowClient defines the interface for scheduling and managing workflows
type WorkflowClient interface {
	ScheduleReplicationTask(ctx context.Context, taskID int64, scheduleExpression string) (string, error)
	CancelWorkflow(ctx context.Context, workflowID string) error
}

// WorkflowClientImpl is a global variable to hold the workflow client implementation
var WorkflowClientImpl WorkflowClient

// StartReplicationTask starts a replication task using Temporal
func (s *service) StartReplicationTask(ctx context.Context, taskID int64) (string, error) {
	// Check if the task exists
	task, err := s.GetReplicationTask(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve task %d: %w", taskID, err)
	}

	if WorkflowClientImpl == nil {
		// For development/testing without Temporal, just return a success message
		fmt.Printf("Development mode: would start replication task %d (WorkflowClient not available)\n", taskID)
		return fmt.Sprintf("mock-workflow-%d", time.Now().Unix()), nil
	}

	// Use our workflow client to schedule the task
	workflowID, err := WorkflowClientImpl.ScheduleReplicationTask(ctx, taskID, task.Schedule)
	if err != nil {
		return "", fmt.Errorf("failed to schedule replication task %d: %w", taskID, err)
	}

	// In a real implementation, we might store this workflowID in the database
	return workflowID, nil
}

// StopReplicationTask stops a running replication task
func (s *service) StopReplicationTask(ctx context.Context, taskID int64) error {
	// In a real implementation, this would look up the workflow ID for this task
	workflowID := fmt.Sprintf("replication-task-%d", taskID)

	if WorkflowClientImpl == nil {
		// For development/testing without Temporal
		fmt.Printf("Development mode: would stop replication task %d (WorkflowClient not available)\n", taskID)
		return nil
	}

	// Cancel the workflow
	return WorkflowClientImpl.CancelWorkflow(ctx, workflowID)
}

// GetReplicationTaskStatus gets the status of a replication task
func (s *service) GetReplicationTaskStatus(ctx context.Context, taskID int64) (string, error) {
	// In a real implementation, this might:
	// 1. Look up the Temporal workflow ID associated with this task
	// 2. Query the workflow's state
	// 3. Return a user-friendly status

	// For now, just return a placeholder
	return "unknown", nil
}

// ListReplicationRuns lists all runs for a specific replication task
func (s *service) ListReplicationRuns(ctx context.Context, taskID int64) ([]*data.ReplicationRun, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service requires an initialized repository")
	}
	return s.repo.ListReplicationRunsForTask(ctx, taskID)
}

// GetReplicationRunDetails gets details of a specific replication run
func (s *service) GetReplicationRunDetails(ctx context.Context, runID int64) (*data.ReplicationRun, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service requires an initialized repository")
	}
	return s.repo.GetReplicationRun(ctx, runID)
}

// CreateReplicationRun calls the repository to create a new run record.
func (s *service) CreateReplicationRun(ctx context.Context, run *data.ReplicationRun) (int64, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("service requires an initialized repository")
	}
	return s.repo.CreateReplicationRun(ctx, run)
}

// UpdateReplicationRunStatus calls the repository to update a run's status.
func (s *service) UpdateReplicationRunStatus(ctx context.Context, id int64, status string, errorDetails string, endTime *time.Time) error {
	if s.repo == nil {
		return fmt.Errorf("service requires an initialized repository")
	}
	return s.repo.UpdateReplicationRunStatus(ctx, id, status, errorDetails, endTime)
}
