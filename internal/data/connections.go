package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateConnection inserts a new connection record into the database.
// It returns the newly created connection's ID.
func (db *DB) CreateConnection(ctx context.Context, conn *Connection) (int64, error) {
	if db == nil || db.SQL == nil {
		return 0, fmt.Errorf("database connection is not initialized")
	}

	// SQL Server syntax for inserting and returning the ID
	query := `
		INSERT INTO Connections (Name, Type, ConnectionString, CreatedAt, UpdatedAt)
		OUTPUT INSERTED.ID
		VALUES (@p1, @p2, @p3, @p4, @p5);`

	now := time.Now()
	var insertedID int64

	err := db.SQL.QueryRowContext(ctx, query,
		conn.Name, conn.Type, conn.ConnectionString, now, now,
	).Scan(&insertedID)

	if err != nil {
		return 0, fmt.Errorf("error creating connection: %w", err)
	}

	conn.ID = insertedID
	conn.CreatedAt = now
	conn.UpdatedAt = now
	return insertedID, nil
}

// GetConnection retrieves a specific connection by its ID.
func (db *DB) GetConnection(ctx context.Context, id int64) (*Connection, error) {
	if db == nil || db.SQL == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
		SELECT ID, Name, Type, ConnectionString, CreatedAt, UpdatedAt
		FROM Connections
		WHERE ID = @p1;`

	row := db.SQL.QueryRowContext(ctx, query, id)
	var conn Connection

	err := row.Scan(
		&conn.ID,
		&conn.Name,
		&conn.Type,
		&conn.ConnectionString,
		&conn.CreatedAt,
		&conn.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows // Return standard error for not found
		}
		return nil, fmt.Errorf("error getting connection %d: %w", id, err)
	}

	return &conn, nil
}

// ListConnections retrieves all connection records from the database.
func (db *DB) ListConnections(ctx context.Context) ([]*Connection, error) {
	if db == nil || db.SQL == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
		SELECT ID, Name, Type, ConnectionString, CreatedAt, UpdatedAt
		FROM Connections
		ORDER BY Name;` // Or order by ID, CreatedAt, etc.

	rows, err := db.SQL.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing connections: %w", err)
	}
	defer rows.Close()

	connections := make([]*Connection, 0)
	for rows.Next() {
		var conn Connection
		if err := rows.Scan(
			&conn.ID,
			&conn.Name,
			&conn.Type,
			&conn.ConnectionString,
			&conn.CreatedAt,
			&conn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning connection row: %w", err) // Return on first scan error
		}
		connections = append(connections, &conn)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating connection rows: %w", err)
	}

	return connections, nil
}

// UpdateConnection updates an existing connection record in the database.
func (db *DB) UpdateConnection(ctx context.Context, conn *Connection) error {
	if db == nil || db.SQL == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `
		UPDATE Connections
		SET Name = @p1, Type = @p2, ConnectionString = @p3, UpdatedAt = @p4
		WHERE ID = @p5;`

	now := time.Now()
	result, err := db.SQL.ExecContext(ctx, query,
		conn.Name, conn.Type, conn.ConnectionString, now, conn.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating connection %d: %w", conn.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This error is less critical, maybe just log it
		return fmt.Errorf("error getting rows affected after update for connection %d: %w", conn.ID, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Use standard error if no rows were updated (implies ID not found)
	}

	conn.UpdatedAt = now // Update the struct's timestamp
	return nil
}

// DeleteConnection removes a connection record from the database by its ID.
func (db *DB) DeleteConnection(ctx context.Context, id int64) error {
	if db == nil || db.SQL == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `DELETE FROM Connections WHERE ID = @p1;`

	result, err := db.SQL.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting connection %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected after delete for connection %d: %w", id, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Use standard error if no rows were deleted (implies ID not found)
	}

	return nil
}
