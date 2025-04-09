package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateBenthosConfig inserts a new Benthos configuration record.
func (db *DB) CreateBenthosConfig(ctx context.Context, config *BenthosConfiguration) (int64, error) {
	if db == nil || db.SQL == nil {
		return 0, fmt.Errorf("database connection is not initialized")
	}

	query := `
		INSERT INTO BenthosConfigurations (Name, Configuration, CreatedAt, UpdatedAt)
		VALUES ($1, $2, $3, $4)
		RETURNING ID;`

	now := time.Now()
	var insertedID int64

	err := db.SQL.QueryRowContext(ctx, query,
		config.Name, config.Configuration, now, now,
	).Scan(&insertedID)

	if err != nil {
		return 0, fmt.Errorf("error creating benthos config: %w", err)
	}

	config.ID = insertedID
	config.CreatedAt = now
	config.UpdatedAt = now
	return insertedID, nil
}

// GetBenthosConfig retrieves a specific Benthos configuration by its ID.
func (db *DB) GetBenthosConfig(ctx context.Context, id int64) (*BenthosConfiguration, error) {
	if db == nil || db.SQL == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `SELECT ID, Name, Configuration, CreatedAt, UpdatedAt FROM BenthosConfigurations WHERE ID = $1;`

	row := db.SQL.QueryRowContext(ctx, query, id)
	var config BenthosConfiguration

	err := row.Scan(
		&config.ID,
		&config.Name,
		&config.Configuration,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("error getting benthos config %d: %w", id, err)
	}

	return &config, nil
}

// ListBenthosConfigs retrieves all Benthos configuration records.
func (db *DB) ListBenthosConfigs(ctx context.Context) ([]*BenthosConfiguration, error) {
	if db == nil || db.SQL == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `SELECT ID, Name, Configuration, CreatedAt, UpdatedAt FROM BenthosConfigurations ORDER BY Name;`

	rows, err := db.SQL.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing benthos configs: %w", err)
	}
	defer rows.Close()

	configs := make([]*BenthosConfiguration, 0)
	for rows.Next() {
		var config BenthosConfiguration
		if err := rows.Scan(
			&config.ID,
			&config.Name,
			&config.Configuration,
			&config.CreatedAt,
			&config.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning benthos config row: %w", err)
		}
		configs = append(configs, &config)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating benthos config rows: %w", err)
	}

	return configs, nil
}

// UpdateBenthosConfig updates an existing Benthos configuration record.
func (db *DB) UpdateBenthosConfig(ctx context.Context, config *BenthosConfiguration) error {
	if db == nil || db.SQL == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `UPDATE BenthosConfigurations SET Name = $1, Configuration = $2, UpdatedAt = $3 WHERE ID = $4;`

	now := time.Now()
	result, err := db.SQL.ExecContext(ctx, query,
		config.Name, config.Configuration, now, config.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating benthos config %d: %w", config.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected for config %d: %w", config.ID, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	config.UpdatedAt = now
	return nil
}

// DeleteBenthosConfig removes a Benthos configuration record by its ID.
func (db *DB) DeleteBenthosConfig(ctx context.Context, id int64) error {
	if db == nil || db.SQL == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `DELETE FROM BenthosConfigurations WHERE ID = $1;`

	result, err := db.SQL.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting benthos config %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected for config %d: %w", id, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
