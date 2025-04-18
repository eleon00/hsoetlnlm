package service

import (
	"context"
	"time"

	"github.com/eleon00/hsoetlnlm/internal/data"
	// Add other necessary imports like models, etc. later
)

// Service defines the interface for the application's business logic.
type Service interface {
	// HealthCheck() error // Example method

	// Connection methods
	CreateConnection(ctx context.Context, conn *data.Connection) (int64, error)
	GetConnection(ctx context.Context, id int64) (*data.Connection, error)
	ListConnections(ctx context.Context) ([]*data.Connection, error)
	UpdateConnection(ctx context.Context, conn *data.Connection) error
	DeleteConnection(ctx context.Context, id int64) error

	// ReplicationTask methods
	CreateReplicationTask(ctx context.Context, task *data.ReplicationTask) (int64, error)
	GetReplicationTask(ctx context.Context, id int64) (*data.ReplicationTask, error)
	ListReplicationTasks(ctx context.Context) ([]*data.ReplicationTask, error)
	UpdateReplicationTask(ctx context.Context, task *data.ReplicationTask) error
	DeleteReplicationTask(ctx context.Context, id int64) error

	// BenthosConfiguration methods
	CreateBenthosConfig(ctx context.Context, config *data.BenthosConfiguration) (int64, error)
	GetBenthosConfig(ctx context.Context, id int64) (*data.BenthosConfiguration, error)
	ListBenthosConfigs(ctx context.Context) ([]*data.BenthosConfiguration, error)
	UpdateBenthosConfig(ctx context.Context, config *data.BenthosConfiguration) error
	DeleteBenthosConfig(ctx context.Context, id int64) error

	// Replication execution methods (using Temporal)
	StartReplicationTask(ctx context.Context, taskID int64) (string, error)
	StopReplicationTask(ctx context.Context, taskID int64) error
	GetReplicationTaskStatus(ctx context.Context, taskID int64) (string, error)
	ListReplicationRuns(ctx context.Context, taskID int64) ([]*data.ReplicationRun, error)
	GetReplicationRunDetails(ctx context.Context, runID int64) (*data.ReplicationRun, error)
	CreateReplicationRun(ctx context.Context, run *data.ReplicationRun) (int64, error)
	UpdateReplicationRunStatus(ctx context.Context, id int64, status string, errorDetails string, endTime *time.Time) error

	// ... other business logic methods
}

// service implements the Service interface.
type service struct {
	repo data.Repository // Dependency on the data layer
	// Add other dependencies like Temporal client, config, logger etc.
}

// NewService creates a new service instance.
func NewService(repo data.Repository) Service {
	return &service{
		repo: repo,
	}
}

// Example implementation (will be expanded later)
// func (s *service) HealthCheck() error {
// 	 // Potentially check repository health or other dependencies
// 	 return nil
// }
