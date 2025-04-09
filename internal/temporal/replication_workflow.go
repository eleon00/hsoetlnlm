package temporal

import (
	"fmt"
	"time"

	"github.com/eleon00/hsoetlnlm/internal/data"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ReplicationWorkflow implements data replication using Temporal
// It orchestrates the entire process from loading task config to running Benthos
func ReplicationWorkflow(ctx workflow.Context, taskID int64) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting replication workflow", "taskID", taskID)

	// Workflow options
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute * 5,
		MaximumAttempts:    3,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// Initialize workflow parameters
	params := WorkflowParams{
		TaskID:    taskID,
		StartTime: workflow.Now(ctx),
		State:     ReplicationWorkflowStateInitialized,
	}

	// Step 1: Create a replication run record in the database
	var run *data.ReplicationRun
	err := workflow.ExecuteActivity(ctx, "CreateReplicationRun", taskID).Get(ctx, &run)
	if err != nil {
		return handleWorkflowError(ctx, &params, "Failed to create replication run", err)
	}
	params.ReplicationRunID = run.ID
	params.State = ReplicationWorkflowStateLoading

	// Step 2: Load the replication task configuration
	var task *data.ReplicationTask
	err = workflow.ExecuteActivity(ctx, "LoadReplicationTask", taskID).Get(ctx, &task)
	if err != nil {
		return handleWorkflowError(ctx, &params, "Failed to load replication task", err)
	}
	params.State = ReplicationWorkflowStateGeneratingConfig

	// Step 3: Generate Benthos configuration
	var benthosConfig *data.BenthosConfiguration
	err = workflow.ExecuteActivity(ctx, "GenerateBenthosConfig", task).Get(ctx, &benthosConfig)
	if err != nil {
		return handleWorkflowError(ctx, &params, "Failed to generate Benthos config", err)
	}
	if benthosConfig != nil && benthosConfig.ID > 0 {
		params.BenthosConfigID = &benthosConfig.ID
	}
	params.State = ReplicationWorkflowStateStartingBenthos

	// Step 4: Start Benthos pipeline
	var processID string
	err = workflow.ExecuteActivity(ctx, "StartBenthosPipeline", benthosConfig).Get(ctx, &processID)
	if err != nil {
		return handleWorkflowError(ctx, &params, "Failed to start Benthos pipeline", err)
	}
	params.BenthosProcessID = processID
	params.State = ReplicationWorkflowStateRunning

	// Step 5: Monitor Benthos pipeline execution
	// In a real implementation, this would be a loop with a selector to handle cancellation
	var completed bool
	monitorCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour, // Longer timeout for monitoring
		HeartbeatTimeout:    time.Minute,
	})
	err = workflow.ExecuteActivity(monitorCtx, "MonitorBenthosPipeline", processID).Get(ctx, &completed)
	if err != nil {
		// Try to stop the pipeline before returning error
		_ = workflow.ExecuteActivity(ctx, "StopBenthosPipeline", processID).Get(ctx, nil)
		return handleWorkflowError(ctx, &params, "Failed to monitor Benthos pipeline", err)
	}

	// Step 6: Update run status to completed
	now := workflow.Now(ctx)
	params.EndTime = &now
	params.State = ReplicationWorkflowStateCompleted
	err = workflow.ExecuteActivity(ctx, "UpdateReplicationRunStatus",
		params.ReplicationRunID, "success", "").Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to update replication run status", "error", err)
		// Continue even though update failed
	}

	logger.Info("Replication workflow completed successfully", "taskID", taskID)
	return nil
}

// handleWorkflowError handles errors in the workflow, updates the run status, and returns the error
func handleWorkflowError(ctx workflow.Context, params *WorkflowParams, msg string, err error) error {
	logger := workflow.GetLogger(ctx)
	logger.Error(msg, "error", err, "taskID", params.TaskID)

	params.State = ReplicationWorkflowStateFailed
	params.ErrorMessage = fmt.Sprintf("%s: %v", msg, err)

	now := workflow.Now(ctx)
	params.EndTime = &now

	// Try to update the run status
	updateErr := workflow.ExecuteActivity(ctx, "UpdateReplicationRunStatus",
		params.ReplicationRunID, "failed", params.ErrorMessage).Get(ctx, nil)
	if updateErr != nil {
		logger.Error("Failed to update replication run status", "error", updateErr)
		// Continue anyway
	}

	return fmt.Errorf("%s: %w", msg, err)
}
