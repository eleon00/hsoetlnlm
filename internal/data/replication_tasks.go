package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateReplicationTask inserts a new replication task record into the database.
func (db *DB) CreateReplicationTask(ctx context.Context, task *ReplicationTask) (int64, error) {
	if db == nil || db.SQL == nil {
		return 0, fmt.Errorf("database connection is not initialized")
	}

	query := `
		INSERT INTO ReplicationTasks (Name, SourceConnectionID, TargetConnectionID, Schedule, DataSelectionCriteria, TransformationRules, Status, CreatedAt, UpdatedAt)
		OUTPUT INSERTED.ID
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9);`

	now := time.Now()
	var insertedID int64

	err := db.SQL.QueryRowContext(ctx, query,
		task.Name,
		task.SourceConnectionID,
		task.TargetConnectionID,
		task.Schedule,              // Use value directly
		task.DataSelectionCriteria, // Use value directly
		task.TransformationRules,   // Use value directly
		"inactive",                 // Default status on creation
		now,
		now,
	).Scan(&insertedID)

	if err != nil {
		return 0, fmt.Errorf("error creating replication task: %w", err)
	}

	task.ID = insertedID
	task.Status = "inactive"
	task.CreatedAt = now
	task.UpdatedAt = now
	return insertedID, nil
}

// GetReplicationTask retrieves a specific replication task by its ID.
func (db *DB) GetReplicationTask(ctx context.Context, id int64) (*ReplicationTask, error) {
	if db == nil || db.SQL == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
		SELECT ID, Name, SourceConnectionID, TargetConnectionID, Schedule, DataSelectionCriteria, TransformationRules, TemporalWorkflowID, Status, CreatedAt, UpdatedAt
		FROM ReplicationTasks
		WHERE ID = @p1;`

	row := db.SQL.QueryRowContext(ctx, query, id)
	var task ReplicationTask
	// Use sql.NullString for potentially nullable string fields
	var schedule, dataSelection, transformRules, temporalWorkflowID sql.NullString

	err := row.Scan(
		&task.ID,
		&task.Name,
		&task.SourceConnectionID,
		&task.TargetConnectionID,
		&schedule,
		&dataSelection,
		&transformRules,
		&temporalWorkflowID,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("error getting replication task %d: %w", id, err)
	}

	// Assign values from sql.NullString if they are valid
	if schedule.Valid {
		task.Schedule = schedule.String
	}
	if dataSelection.Valid {
		task.DataSelectionCriteria = dataSelection.String
	}
	if transformRules.Valid {
		task.TransformationRules = transformRules.String
	}
	if temporalWorkflowID.Valid {
		task.TemporalWorkflowID = temporalWorkflowID.String
	}

	return &task, nil
}

// ListReplicationTasks retrieves all replication task records from the database.
func (db *DB) ListReplicationTasks(ctx context.Context) ([]*ReplicationTask, error) {
	if db == nil || db.SQL == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
		SELECT ID, Name, SourceConnectionID, TargetConnectionID, Schedule, DataSelectionCriteria, TransformationRules, TemporalWorkflowID, Status, CreatedAt, UpdatedAt
		FROM ReplicationTasks
		ORDER BY Name;`

	rows, err := db.SQL.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing replication tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*ReplicationTask, 0)
	for rows.Next() {
		var task ReplicationTask
		var schedule, dataSelection, transformRules, temporalWorkflowID sql.NullString

		if err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.SourceConnectionID,
			&task.TargetConnectionID,
			&schedule,
			&dataSelection,
			&transformRules,
			&temporalWorkflowID,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning replication task row: %w", err)
		}

		if schedule.Valid {
			task.Schedule = schedule.String
		}
		if dataSelection.Valid {
			task.DataSelectionCriteria = dataSelection.String
		}
		if transformRules.Valid {
			task.TransformationRules = transformRules.String
		}
		if temporalWorkflowID.Valid {
			task.TemporalWorkflowID = temporalWorkflowID.String
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating replication task rows: %w", err)
	}

	return tasks, nil
}

// UpdateReplicationTask updates an existing replication task record.
// Note: This implementation updates most fields. You might want more granular updates.
func (db *DB) UpdateReplicationTask(ctx context.Context, task *ReplicationTask) error {
	if db == nil || db.SQL == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `
		UPDATE ReplicationTasks
		SET Name = @p1, SourceConnectionID = @p2, TargetConnectionID = @p3,
		    Schedule = @p4, DataSelectionCriteria = @p5, TransformationRules = @p6,
		    TemporalWorkflowID = @p7, Status = @p8, UpdatedAt = @p9
		WHERE ID = @p10;`

	now := time.Now()
	result, err := db.SQL.ExecContext(ctx, query,
		task.Name,
		task.SourceConnectionID,
		task.TargetConnectionID,
		sql.NullString{String: task.Schedule, Valid: task.Schedule != ""}, // Handle potential empty strings
		sql.NullString{String: task.DataSelectionCriteria, Valid: task.DataSelectionCriteria != ""},
		sql.NullString{String: task.TransformationRules, Valid: task.TransformationRules != ""},
		sql.NullString{String: task.TemporalWorkflowID, Valid: task.TemporalWorkflowID != ""},
		task.Status,
		now,
		task.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating replication task %d: %w", task.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected for task %d: %w", task.ID, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // No rows updated means ID not found
	}

	task.UpdatedAt = now
	return nil
}

// DeleteReplicationTask removes a replication task record by its ID.
func (db *DB) DeleteReplicationTask(ctx context.Context, id int64) error {
	if db == nil || db.SQL == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `DELETE FROM ReplicationTasks WHERE ID = @p1;`

	result, err := db.SQL.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting replication task %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected for task %d: %w", id, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // ID not found
	}

	return nil
}
