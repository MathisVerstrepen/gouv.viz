package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQL string

func configureDatabase(db *sql.DB) error {
	statements := []string{
		"PRAGMA journal_mode = OFF",
		"PRAGMA synchronous = OFF",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA foreign_keys = OFF",
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("configure sqlite database: %w", err)
		}
	}
	return nil
}

func createSchema(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("create sqlite schema: %w", err)
	}
	return nil
}

func validateDatabase(db *sql.DB) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open sqlite validation connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enable sqlite foreign keys: %w", err)
	}

	rows, err := conn.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("check sqlite foreign keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		var rowID sql.NullInt64
		var parent string
		var fkID int
		if err := rows.Scan(&table, &rowID, &parent, &fkID); err != nil {
			return fmt.Errorf("scan foreign key violation: %w", err)
		}
		if rowID.Valid {
			return fmt.Errorf("foreign key violation: table %q rowid %d references %q (fk %d)", table, rowID.Int64, parent, fkID)
		}
		return fmt.Errorf("foreign key violation: table %q references %q (fk %d)", table, parent, fkID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate foreign key violations: %w", err)
	}

	return nil
}
