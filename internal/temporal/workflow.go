package temporal

import (
	"context"
	"time"

	"github.com/eleon00/hsoetlnlm/internal/data"
)

// ReplicationWorkflowState represents the current state of a replication workflow
type ReplicationWorkflowState string

const (
	// ReplicationWorkflowStateInitialized is the initial state when a workflow is created
	ReplicationWorkflowStateInitialized ReplicationWorkflowState = "initialized"
	// ReplicationWorkflowStateLoading is the state when loading task configuration
	ReplicationWorkflowStateLoading ReplicationWorkflowState = "loading"
	// ReplicationWorkflowStateGeneratingConfig is the state when generating Benthos config
	ReplicationWorkflowStateGeneratingConfig ReplicationWorkflowState = "generating_config"
	// ReplicationWorkflowStateStartingBenthos is the state when starting Benthos pipeline
	ReplicationWorkflowStateStartingBenthos ReplicationWorkflowState = "starting_benthos"
	// ReplicationWorkflowStateRunning is the state when Benthos pipeline is running
	ReplicationWorkflowStateRunning ReplicationWorkflowState = "running"
	// ReplicationWorkflowStateCompleted is the state when replication completes successfully
	ReplicationWorkflowStateCompleted ReplicationWorkflowState = "completed"
	// ReplicationWorkflowStateFailed is the state when replication fails
	ReplicationWorkflowStateFailed ReplicationWorkflowState = "failed"
)

// WorkflowParams contains parameters needed by replication workflows
type WorkflowParams struct {
	TaskID           int64                    `json:"task_id"`
	StartTime        time.Time                `json:"start_time"`
	EndTime          *time.Time               `json:"end_time,omitempty"`
	State            ReplicationWorkflowState `json:"state"`
	ErrorMessage     string                   `json:"error_message,omitempty"`
	BenthosConfigID  *int64                   `json:"benthos_config_id,omitempty"`
	BenthosProcessID string                   `json:"benthos_process_id,omitempty"`
	ReplicationRunID int64                    `json:"replication_run_id,omitempty"`
}

// Activities interface defines activity methods used by replication workflows
type Activities interface {
	// LoadReplicationTask loads the task configuration from the database
	LoadReplicationTask(ctx context.Context, taskID int64) (*data.ReplicationTask, error)

	// CreateReplicationRun creates a new replication run record in the database
	CreateReplicationRun(ctx context.Context, taskID int64) (*data.ReplicationRun, error)

	// GenerateBenthosConfig generates a Benthos configuration for the task
	GenerateBenthosConfig(ctx context.Context, task *data.ReplicationTask) (*data.BenthosConfiguration, error)

	// StartBenthosPipeline starts a Benthos pipeline with the given configuration
	StartBenthosPipeline(ctx context.Context, config *data.BenthosConfiguration) (string, error)

	// MonitorBenthosPipeline checks the status of a running Benthos pipeline
	MonitorBenthosPipeline(ctx context.Context, processID string) (bool, error)

	// StopBenthosPipeline stops a running Benthos pipeline
	StopBenthosPipeline(ctx context.Context, processID string) error

	// UpdateReplicationRunStatus updates the status of a replication run
	UpdateReplicationRunStatus(ctx context.Context, runID int64, status string, errorMsg string) error
}
