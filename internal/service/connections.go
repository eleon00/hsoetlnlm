package service

import (
	"context"
	"fmt"

	"github.com/eleon00/hsoetlnlm/internal/data"
)

// CreateConnection handles the business logic for creating a connection.
func (s *service) CreateConnection(ctx context.Context, conn *data.Connection) (int64, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Add validation logic for the connection data
	fmt.Printf("Service: Calling repo.CreateConnection for '%s'\n", conn.Name)
	return s.repo.CreateConnection(ctx, conn)
}

// GetConnection handles the business logic for retrieving a connection by ID.
func (s *service) GetConnection(ctx context.Context, id int64) (*data.Connection, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service requires an initialized repository")
	}
	fmt.Printf("Service: Calling repo.GetConnection for ID %d\n", id)
	return s.repo.GetConnection(ctx, id)
}

// ListConnections handles the business logic for listing all connections.
func (s *service) ListConnections(ctx context.Context) ([]*data.Connection, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service requires an initialized repository")
	}
	fmt.Println("Service: Calling repo.ListConnections")
	return s.repo.ListConnections(ctx)
}

// UpdateConnection handles the business logic for updating a connection.
func (s *service) UpdateConnection(ctx context.Context, conn *data.Connection) error {
	if s.repo == nil {
		return fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Add validation logic for the connection data
	// TODO: Potentially check if the connection exists before updating
	fmt.Printf("Service: Calling repo.UpdateConnection for ID %d\n", conn.ID)
	return s.repo.UpdateConnection(ctx, conn)
}

// DeleteConnection handles the business logic for deleting a connection by ID.
func (s *service) DeleteConnection(ctx context.Context, id int64) error {
	if s.repo == nil {
		return fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Add logic to check if the connection is in use by any tasks before deleting
	fmt.Printf("Service: Calling repo.DeleteConnection for ID %d\n", id)
	return s.repo.DeleteConnection(ctx, id)
}
