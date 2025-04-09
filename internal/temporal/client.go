package temporal

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/eleon00/hsoetlnlm/internal/service"
)

// Ensure our Client implements the WorkflowClient interface
var _ service.WorkflowClient = (*Client)(nil)

// Client wraps the Temporal client and provides application-specific methods
type Client struct {
	tc client.Client // The actual Temporal client
}

// ClientOptions defines options for creating a new Temporal client
type ClientOptions struct {
	HostPort    string // Temporal server address (default: "localhost:7233")
	Namespace   string // Temporal namespace (default: "default")
	ServiceName string // Name to identify this service (default: "hsoetlnlm")
}

// NewClient creates a new Temporal client with the provided options
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		opts = &ClientOptions{}
	}

	// Set defaults for empty options
	if opts.HostPort == "" {
		opts.HostPort = "localhost:7233"
	}
	if opts.Namespace == "" {
		opts.Namespace = "default"
	}
	if opts.ServiceName == "" {
		opts.ServiceName = "hsoetlnlm"
	}

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort:  opts.HostPort,
		Namespace: opts.Namespace,
		Identity:  opts.ServiceName,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	return &Client{tc: c}, nil
}

// Close closes the Temporal client connection
func (c *Client) Close() {
	if c.tc != nil {
		c.tc.Close()
	}
}

// GetTemporalClient returns the underlying Temporal client
func (c *Client) GetTemporalClient() client.Client {
	return c.tc
}

// ExecuteWorkflow wraps the Temporal client's ExecuteWorkflow method with common options
func (c *Client) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	return c.tc.ExecuteWorkflow(ctx, options, workflow, args...)
}

// ScheduleReplicationTask starts or schedules a replication task workflow
func (c *Client) ScheduleReplicationTask(ctx context.Context, taskID int64, scheduleExpression string) (string, error) {
	// Default options
	options := client.StartWorkflowOptions{
		ID:                  fmt.Sprintf("replication-task-%d", taskID),
		TaskQueue:           "replication-tasks",
		WorkflowRunTimeout:  time.Hour * 24, // 24-hour timeout for long-running workflows
		WorkflowTaskTimeout: time.Minute * 10,
		// Note: In a real implementation, we would configure scheduling based on the
		// scheduleExpression parameter (like a cron expression). For now, we just
		// execute immediately.
	}

	// Execute the workflow
	run, err := c.ExecuteWorkflow(ctx, options, ReplicationWorkflow, taskID)
	if err != nil {
		return "", fmt.Errorf("failed to start replication workflow for task %d: %w", taskID, err)
	}

	// Return the workflow ID and run ID
	return run.GetID(), nil
}

// CancelWorkflow cancels a running workflow
func (c *Client) CancelWorkflow(ctx context.Context, workflowID string) error {
	// In a real implementation, we might need to look up the current run ID
	// For now, we'll use an empty string which cancels the current run
	return c.tc.CancelWorkflow(ctx, workflowID, "")
}
