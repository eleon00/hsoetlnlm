package temporal

import (
	"fmt"

	"github.com/eleon00/hsoetlnlm/internal/service"
	"go.temporal.io/sdk/worker"
)

// WorkerOptions defines options for creating a new worker
type WorkerOptions struct {
	ServiceName string // Name to identify this service
	TaskQueue   string // Name of the task queue to listen on
}

// Worker wraps the Temporal worker
type Worker struct {
	client *Client
	worker worker.Worker
}

// NewWorker creates a new Temporal worker with the provided options
func NewWorker(client *Client, svc service.Service, opts *WorkerOptions) (*Worker, error) {
	if client == nil {
		return nil, fmt.Errorf("temporal client is required")
	}
	if opts == nil {
		opts = &WorkerOptions{}
	}

	// Set defaults for empty options
	if opts.TaskQueue == "" {
		opts.TaskQueue = "replication-tasks"
	}

	// Create a Temporal worker
	w := worker.New(client.GetTemporalClient(), opts.TaskQueue, worker.Options{})

	// Register workflow handlers
	w.RegisterWorkflow(ReplicationWorkflow)

	// Register activity handlers
	activities := NewActivities(svc)
	w.RegisterActivity(activities)

	return &Worker{
		client: client,
		worker: w,
	}, nil
}

// Start begins listening for tasks and executing workflows/activities
func (w *Worker) Start() error {
	return w.worker.Start()
}

// Stop gracefully shuts down the worker
func (w *Worker) Stop() {
	w.worker.Stop()
}
