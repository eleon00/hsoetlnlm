package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Import PostgreSQL driver
	_ "github.com/lib/pq"
)

// Repository defines the interface for database operations.
// This allows for mocking in tests and swapping implementations.
type Repository interface {
	// Connection methods
	CreateConnection(ctx context.Context, conn *Connection) (int64, error)
	GetConnection(ctx context.Context, id int64) (*Connection, error)
	ListConnections(ctx context.Context) ([]*Connection, error)
	UpdateConnection(ctx context.Context, conn *Connection) error
	DeleteConnection(ctx context.Context, id int64) error

	// ReplicationTask methods (Placeholders)
	CreateReplicationTask(ctx context.Context, task *ReplicationTask) (int64, error)
	GetReplicationTask(ctx context.Context, id int64) (*ReplicationTask, error)
	ListReplicationTasks(ctx context.Context) ([]*ReplicationTask, error)
	UpdateReplicationTask(ctx context.Context, task *ReplicationTask) error
	DeleteReplicationTask(ctx context.Context, id int64) error

	// BenthosConfiguration methods (Placeholders)
	CreateBenthosConfig(ctx context.Context, config *BenthosConfiguration) (int64, error)
	GetBenthosConfig(ctx context.Context, id int64) (*BenthosConfiguration, error)
	ListBenthosConfigs(ctx context.Context) ([]*BenthosConfiguration, error)
	UpdateBenthosConfig(ctx context.Context, config *BenthosConfiguration) error
	DeleteBenthosConfig(ctx context.Context, id int64) error

	// ReplicationRun methods
	CreateReplicationRun(ctx context.Context, run *ReplicationRun) (int64, error)
	GetReplicationRun(ctx context.Context, id int64) (*ReplicationRun, error)
	ListReplicationRunsForTask(ctx context.Context, taskID int64) ([]*ReplicationRun, error)
	UpdateReplicationRunStatus(ctx context.Context, id int64, status string, errorDetails string, endTime *time.Time) error

	// Placeholder methods for other resources
	// GetReplicationRun(ctx context.Context, id int64) (*ReplicationRun, error)
	// ... other CRUD operations for ReplicationTask, ReplicationRun, etc.
}

// DB holds the database connection pool.
type DB struct {
	SQL *sql.DB
}

// NewDB initializes a new database connection pool.
func NewDB(dataSourceName string) (*DB, error) {
	if dataSourceName == "" {
		return nil, fmt.Errorf("database data source name is required")
	}

	// Use "postgres" as the driver name
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Ping verifies the connection is alive.
	if err = db.Ping(); err != nil {
		db.Close() // Close if ping fails
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Configure connection pool settings (example)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{SQL: db}, nil
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	if db.SQL != nil {
		return db.SQL.Close()
	}
	return nil
}
