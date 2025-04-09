package data

import (
	"time"
)

// Connection represents the Connections table.
// Stores details for connecting to source or target systems.
type Connection struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"` // e.g., 'oracle', 'sqlserver', 's3', 'bigquery', 'snowflake', 'localfile'
	ConnectionString string    `json:"connection_string"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ReplicationTask represents the ReplicationTasks table.
// Defines a data replication job.
type ReplicationTask struct {
	ID                    int64     `json:"id"`
	Name                  string    `json:"name"`
	SourceConnectionID    int64     `json:"source_connection_id"`
	TargetConnectionID    int64     `json:"target_connection_id"`
	Schedule              string    `json:"schedule,omitempty"` // e.g., cron expression
	DataSelectionCriteria string    `json:"data_selection_criteria,omitempty"`
	TransformationRules   string    `json:"transformation_rules,omitempty"` // e.g., Benthos Bloblang
	TemporalWorkflowID    string    `json:"temporal_workflow_id,omitempty"`
	Status                string    `json:"status"` // e.g., 'active', 'inactive', 'failed'
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// ReplicationRun represents the ReplicationRuns table.
// Stores the history and status of a specific execution of a ReplicationTask.
type ReplicationRun struct {
	ID                int64      `json:"id"`
	ReplicationTaskID int64      `json:"replication_task_id"`
	StartTime         time.Time  `json:"start_time"`
	EndTime           *time.Time `json:"end_time,omitempty"` // Pointer allows for NULL values
	Status            string     `json:"status"`             // e.g., 'running', 'success', 'failed'
	ErrorDetails      string     `json:"error_details,omitempty"`
	TemporalRunID     string     `json:"temporal_run_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// BenthosConfiguration represents the BenthosConfigurations table.
// Stores reusable Benthos pipeline configurations.
type BenthosConfiguration struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Configuration string    `json:"configuration"` // Benthos YAML or JSON
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
