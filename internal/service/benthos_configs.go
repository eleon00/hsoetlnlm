package service

import (
	"context"
	"fmt"

	"github.com/eleon00/hsoetlnlm/internal/data"
)

// CreateBenthosConfig handles the business logic for creating a Benthos config.
func (s *service) CreateBenthosConfig(ctx context.Context, config *data.BenthosConfiguration) (int64, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Add validation (e.g., check if Benthos config syntax is valid?)
	fmt.Printf("Service: Calling repo.CreateBenthosConfig for '%s'\n", config.Name)
	return s.repo.CreateBenthosConfig(ctx, config)
}

// GetBenthosConfig handles the business logic for retrieving a Benthos config by ID.
func (s *service) GetBenthosConfig(ctx context.Context, id int64) (*data.BenthosConfiguration, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service requires an initialized repository")
	}
	fmt.Printf("Service: Calling repo.GetBenthosConfig for ID %d\n", id)
	return s.repo.GetBenthosConfig(ctx, id)
}

// ListBenthosConfigs handles the business logic for listing all Benthos configs.
func (s *service) ListBenthosConfigs(ctx context.Context) ([]*data.BenthosConfiguration, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service requires an initialized repository")
	}
	fmt.Println("Service: Calling repo.ListBenthosConfigs")
	return s.repo.ListBenthosConfigs(ctx)
}

// UpdateBenthosConfig handles the business logic for updating a Benthos config.
func (s *service) UpdateBenthosConfig(ctx context.Context, config *data.BenthosConfiguration) error {
	if s.repo == nil {
		return fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Add validation
	fmt.Printf("Service: Calling repo.UpdateBenthosConfig for ID %d\n", config.ID)
	return s.repo.UpdateBenthosConfig(ctx, config)
}

// DeleteBenthosConfig handles the business logic for deleting a Benthos config by ID.
func (s *service) DeleteBenthosConfig(ctx context.Context, id int64) error {
	if s.repo == nil {
		return fmt.Errorf("service requires an initialized repository")
	}
	// TODO: Check if config is currently used by any ReplicationTasks?
	fmt.Printf("Service: Calling repo.DeleteBenthosConfig for ID %d\n", id)
	return s.repo.DeleteBenthosConfig(ctx, id)
}
