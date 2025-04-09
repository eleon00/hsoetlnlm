package service

import (
	"context"
	"fmt"

	"github.com/eleon00/hsoetlnlm/internal/data"
)

// CreateReplicationTask handles the business logic for creating a replication task.
func (s *service) CreateReplicationTask(ctx context.Context, task *data.ReplicationTask) (int64, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Add validation logic (e.g., check if Source/Target Connection IDs exist)
	fmt.Printf("Service: Calling repo.CreateReplicationTask for '%s'\n", task.Name)
	return s.repo.CreateReplicationTask(ctx, task)
}

// GetReplicationTask handles the business logic for retrieving a replication task by ID.
func (s *service) GetReplicationTask(ctx context.Context, id int64) (*data.ReplicationTask, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service requires an initialized repository")
	}
	fmt.Printf("Service: Calling repo.GetReplicationTask for ID %d\n", id)
	return s.repo.GetReplicationTask(ctx, id)
}

// ListReplicationTasks handles the business logic for listing all replication tasks.
func (s *service) ListReplicationTasks(ctx context.Context) ([]*data.ReplicationTask, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service requires an initialized repository")
	}
	fmt.Println("Service: Calling repo.ListReplicationTasks")
	return s.repo.ListReplicationTasks(ctx)
}

// UpdateReplicationTask handles the business logic for updating a replication task.
func (s *service) UpdateReplicationTask(ctx context.Context, task *data.ReplicationTask) error {
	if s.repo == nil {
		return fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Add validation logic
	// TODO: Potentially check if the task exists before updating
	fmt.Printf("Service: Calling repo.UpdateReplicationTask for ID %d\n", task.ID)
	return s.repo.UpdateReplicationTask(ctx, task)
}

// DeleteReplicationTask handles the business logic for deleting a replication task by ID.
func (s *service) DeleteReplicationTask(ctx context.Context, id int64) error {
	if s.repo == nil {
		return fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Add logic (e.g., stop associated temporal workflow? check run history?)
	fmt.Printf("Service: Calling repo.DeleteReplicationTask for ID %d\n", id)
	return s.repo.DeleteReplicationTask(ctx, id)
}
