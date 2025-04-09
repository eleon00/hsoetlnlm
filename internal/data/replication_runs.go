package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateReplicationRun inserts a new replication run record.
func (db *DB) CreateReplicationRun(ctx context.Context, run *ReplicationRun) (int64, error) {
	if db == nil || db.SQL == nil {
		return 0, fmt.Errorf("database connection is not initialized")
	}

	query := `
		INSERT INTO ReplicationRuns (ReplicationTaskID, StartTime, Status, TemporalRunID, CreatedAt)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ID;`

	now := time.Now()
	var insertedID int64

	err := db.SQL.QueryRowContext(ctx, query,
		run.ReplicationTaskID,
		run.StartTime,
		run.Status,
		sql.NullString{String: run.TemporalRunID, Valid: run.TemporalRunID != ""},
		now,
	).Scan(&insertedID)

	if err != nil {
		return 0, fmt.Errorf("error creating replication run: %w", err)
	}

	run.ID = insertedID
	run.CreatedAt = now
	return insertedID, nil
}

// GetReplicationRun retrieves a specific replication run by its ID.
func (db *DB) GetReplicationRun(ctx context.Context, id int64) (*ReplicationRun, error) {
	if db == nil || db.SQL == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
		SELECT ID, ReplicationTaskID, StartTime, EndTime, Status, ErrorDetails, TemporalRunID, CreatedAt
		FROM ReplicationRuns
		WHERE ID = $1;`

	row := db.SQL.QueryRowContext(ctx, query, id)
	var run ReplicationRun
	var endTime sql.NullTime
	var errorDetails, temporalRunID sql.NullString

	err := row.Scan(
		&run.ID,
		&run.ReplicationTaskID,
		&run.StartTime,
		&endTime,
		&run.Status,
		&errorDetails,
		&temporalRunID,
		&run.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("error getting replication run %d: %w", id, err)
	}

	if endTime.Valid {
		run.EndTime = &endTime.Time
	}
	if errorDetails.Valid {
		run.ErrorDetails = errorDetails.String
	}
	if temporalRunID.Valid {
		run.TemporalRunID = temporalRunID.String
	}

	return &run, nil
}

// ListReplicationRunsForTask retrieves all runs for a specific task ID.
func (db *DB) ListReplicationRunsForTask(ctx context.Context, taskID int64) ([]*ReplicationRun, error) {
	if db == nil || db.SQL == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
		SELECT ID, ReplicationTaskID, StartTime, EndTime, Status, ErrorDetails, TemporalRunID, CreatedAt
		FROM ReplicationRuns
		WHERE ReplicationTaskID = $1
		ORDER BY StartTime DESC;` // Show most recent first

	rows, err := db.SQL.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("error listing runs for task %d: %w", taskID, err)
	}
	defer rows.Close()

	runs := make([]*ReplicationRun, 0)
	for rows.Next() {
		var run ReplicationRun
		var endTime sql.NullTime
		var errorDetails, temporalRunID sql.NullString

		if err := rows.Scan(
			&run.ID,
			&run.ReplicationTaskID,
			&run.StartTime,
			&endTime,
			&run.Status,
			&errorDetails,
			&temporalRunID,
			&run.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning replication run row: %w", err)
		}

		if endTime.Valid {
			run.EndTime = &endTime.Time
		}
		if errorDetails.Valid {
			run.ErrorDetails = errorDetails.String
		}
		if temporalRunID.Valid {
			run.TemporalRunID = temporalRunID.String
		}

		runs = append(runs, &run)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating replication run rows: %w", err)
	}

	return runs, nil
}

// UpdateReplicationRunStatus updates the status, error details, and end time of a run.
func (db *DB) UpdateReplicationRunStatus(ctx context.Context, id int64, status string, errorDetails string, endTime *time.Time) error {
	if db == nil || db.SQL == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `
		UPDATE ReplicationRuns
		SET Status = $1, ErrorDetails = $2, EndTime = $3
		WHERE ID = $4;`

	// Convert *time.Time to sql.NullTime
	var nullEndTime sql.NullTime
	if endTime != nil {
		nullEndTime = sql.NullTime{Time: *endTime, Valid: true}
	}

	result, err := db.SQL.ExecContext(ctx, query,
		status,
		sql.NullString{String: errorDetails, Valid: errorDetails != ""},
		nullEndTime,
		id,
	)

	if err != nil {
		return fmt.Errorf("error updating status for replication run %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected for run %d: %w", id, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // ID not found
	}

	return nil
}
