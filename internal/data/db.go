package data

import (
	"context"
	"database/sql"
	"fmt"

	// Add specific database driver imports here later, e.g.:
	// _ "github.com/lib/pq"
	// Import the SQL Server driver, the blank identifier registers the driver
	_ "github.com/microsoft/go-mssqldb"
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

	// Placeholder methods for other resources
	// GetReplicationRun(ctx context.Context, id int64) (*ReplicationRun, error)
	// ... other CRUD operations for ReplicationTask, ReplicationRun, etc.
}

// DB holds the database connection pool.
type DB struct {
	SQL *sql.DB
}

// NewDB initializes a new database connection pool.
// Configuration (like DSN) will be needed here.
func NewDB(dataSourceName string) (*DB, error) {
	if dataSourceName == "" {
		// In a real app, load this from config
		return nil, fmt.Errorf("database data source name is required")
	}

	// sql.Open doesn't actually create a connection, just prepares.
	// Ensure the driver name matches the imported driver.
	db, err := sql.Open("sqlserver", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Ping verifies the connection is alive.
	if err = db.Ping(); err != nil {
		db.Close() // Close if ping fails
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// TODO: Configure connection pool settings (MaxOpenConns, MaxIdleConns, etc.)

	return &DB{SQL: db}, nil
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	if db.SQL != nil {
		return db.SQL.Close()
	}
	return nil
}
