package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/eleon00/hsoetlnlm/internal/data"
	"github.com/eleon00/hsoetlnlm/internal/service"
)

// ActivitiesImpl implements the Activities interface
type ActivitiesImpl struct {
	svc service.Service // Service layer dependency
}

// NewActivities creates a new activities implementation
func NewActivities(svc service.Service) Activities {
	return &ActivitiesImpl{svc: svc}
}

// LoadReplicationTask loads the replication task configuration from the database
func (a *ActivitiesImpl) LoadReplicationTask(ctx context.Context, taskID int64) (*data.ReplicationTask, error) {
	// Simply call the service to get the task
	task, err := a.svc.GetReplicationTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to load replication task %d: %w", taskID, err)
	}
	return task, nil
}

// CreateReplicationRun creates a new replication run record
func (a *ActivitiesImpl) CreateReplicationRun(ctx context.Context, taskID int64) (*data.ReplicationRun, error) {
	// Call the service to create the run record in the database
	run := &data.ReplicationRun{
		ReplicationTaskID: taskID,
		StartTime:         time.Now(),                              // Activity start time as DB start time
		Status:            string(ReplicationWorkflowStateLoading), // Initial status after creation
		// TemporalRunID can be added later if needed, maybe via update
	}

	newID, err := a.svc.CreateReplicationRun(ctx, run)
	if err != nil {
		return nil, fmt.Errorf("failed to create replication run for task %d: %w", taskID, err)
	}
	run.ID = newID // Update run object with the returned ID

	return run, nil
}

// GenerateBenthosConfig generates a Benthos configuration for the task
func (a *ActivitiesImpl) GenerateBenthosConfig(ctx context.Context, task *data.ReplicationTask) (*data.BenthosConfiguration, error) {
	// This is a simplified implementation
	// In a real scenario, we would:
	// 1. Get source and target connection details from the database using task.SourceConnectionID and task.TargetConnectionID
	// 2. Create appropriate Benthos configuration based on connection types (Oracle, S3, etc.)
	// 3. Apply any transformation rules from task.TransformationRules

	// For now, return a simple example config
	config := &data.BenthosConfiguration{
		Name: fmt.Sprintf("Config for Task %d", task.ID),
		Configuration: `
input:
  generate:
    count: 10
    interval: "1s"
    mapping: 'root = {"id": uuid_v4(), "task_id": "` + fmt.Sprintf("%d", task.ID) + `", "timestamp": now()}'
output:
  stdout: {}
`,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// TODO: Store this config in the database and get an ID
	// Placeholder ID for now
	config.ID = time.Now().Unix()

	return config, nil
}

// StartBenthosPipeline starts a Benthos pipeline with the given configuration
func (a *ActivitiesImpl) StartBenthosPipeline(ctx context.Context, config *data.BenthosConfiguration) (string, error) {
	// In a real implementation, this would:
	// 1. Write the configuration to a file
	// 2. Execute the Benthos process with the configuration file
	// 3. Return a process ID or identifier

	// For now, simulate success
	processID := fmt.Sprintf("benthos-process-%d", time.Now().Unix())
	return processID, nil
}

// MonitorBenthosPipeline checks the status of a running Benthos pipeline
func (a *ActivitiesImpl) MonitorBenthosPipeline(ctx context.Context, processID string) (bool, error) {
	// In a real implementation, this would:
	// 1. Check if the Benthos process is still running
	// 2. Potentially check logs or metrics to ensure it's functioning properly
	// 3. Return completion status

	// For now, simulate success after a short delay
	time.Sleep(time.Second * 2)
	return true, nil
}

// StopBenthosPipeline stops a running Benthos pipeline
func (a *ActivitiesImpl) StopBenthosPipeline(ctx context.Context, processID string) error {
	// In a real implementation, this would:
	// 1. Send a signal to the Benthos process to gracefully shut down
	// 2. If that fails, forcefully terminate the process

	// For now, simulate success
	return nil
}

// UpdateReplicationRunStatus updates the status of a replication run
func (a *ActivitiesImpl) UpdateReplicationRunStatus(ctx context.Context, runID int64, status string, errorMsg string) error {
	// Determine end time based on status
	var endTime *time.Time
	if status == string(ReplicationWorkflowStateCompleted) || status == string(ReplicationWorkflowStateFailed) {
		now := time.Now()
		endTime = &now
	}

	err := a.svc.UpdateReplicationRunStatus(ctx, runID, status, errorMsg, endTime)
	if err != nil {
		// Log the error but don't necessarily fail the activity,
		// as the core workflow might have completed/failed anyway.
		fmt.Printf("Non-fatal error updating replication run %d status: %v\n", runID, err)
	}
	return nil // Return nil to avoid Temporal retrying the activity just for this update failure
}
