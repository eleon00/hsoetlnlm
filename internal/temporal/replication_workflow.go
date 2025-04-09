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

	// Defer cleanup/status update in case of workflow errors/cancellation
	defer func() {
		if ctx.Err() != nil || params.State != ReplicationWorkflowStateCompleted {
			if params.State != ReplicationWorkflowStateFailed {
				params.State = ReplicationWorkflowStateFailed
				if params.ErrorMessage == "" {
					params.ErrorMessage = fmt.Sprintf("Workflow failed or cancelled: %v", ctx.Err())
				}
				now := workflow.Now(ctx)
				params.EndTime = &now
			}
			// Use a disconnected context for the final status update to ensure it runs
			dcCtx, _ := workflow.NewDisconnectedContext(ctx)
			finalUpdateOptions := workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute, // Short timeout for final update
			}
			dcCtx = workflow.WithActivityOptions(dcCtx, finalUpdateOptions)
			err := workflow.ExecuteActivity(dcCtx, "UpdateReplicationRunStatus",
				params.ReplicationRunID, string(params.State), params.ErrorMessage).Get(dcCtx, nil)
			if err != nil {
				logger.Error("Failed to perform final update of replication run status", "error", err, "RunID", params.ReplicationRunID)
			}
		}
	}()

	// Step 1: Create a replication run record in the database
	var run *data.ReplicationRun
	err := workflow.ExecuteActivity(ctx, "CreateReplicationRun", taskID).Get(ctx, &run)
	if err != nil {
		params.ErrorMessage = fmt.Sprintf("Failed to create replication run: %v", err)
		return err // Error handled by defer
	}
	params.ReplicationRunID = run.ID
	params.State = ReplicationWorkflowStateLoading // Run created, now loading task

	// Step 2: Execute Benthos Pipeline Activity
	// This activity now handles loading task, connections, generating config, and running benthos.
	// Use a longer timeout for the Benthos execution itself.
	benthosActivityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 1,   // Example: Allow 1 hour for the pipeline run
		HeartbeatTimeout:    time.Minute * 2, // Send heartbeats during long runs
		RetryPolicy:         retryPolicy,     // Reuse the defined retry policy
	}
	benthosCtx := workflow.WithActivityOptions(ctx, benthosActivityOpts)

	var benthosOutput string
	err = workflow.ExecuteActivity(benthosCtx, "ExecuteBenthosPipelineActivity", taskID, params.ReplicationRunID).Get(benthosCtx, &benthosOutput)
	if err != nil {
		// Error occurred during Benthos execution
		params.ErrorMessage = fmt.Sprintf("Benthos pipeline execution failed: %v", err)
		// Benthos output might contain useful error info
		logger.Error("Benthos execution failed", "error", err, "output", benthosOutput)
		return err // Error handled by defer
	}

	// Benthos pipeline completed successfully (according to the activity)
	logger.Info("Benthos pipeline executed successfully.", "output_snippet", truncateString(benthosOutput, 200))

	// Step 3: Update run status to completed
	now := workflow.Now(ctx)
	params.EndTime = &now
	params.State = ReplicationWorkflowStateCompleted
	err = workflow.ExecuteActivity(ctx, "UpdateReplicationRunStatus",
		params.ReplicationRunID, string(params.State), "").Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to update replication run status to completed", "error", err)
		// Continue even though update failed, workflow itself succeeded.
	}

	logger.Info("Replication workflow completed successfully", "taskID", taskID)
	return nil
}

// Helper function to truncate strings for logging (can be shared or moved)
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
