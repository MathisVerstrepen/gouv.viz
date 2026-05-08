package store

import (
	"context"
	"database/sql"
	"fmt"
)

const expectedSchemaVersion = "1"

var requiredTables = []string{
	"dataset_meta",
	"acteurs",
	"acteur_adresses",
	"organes",
	"mandats",
	"mandat_organes",
	"scrutins",
	"scrutin_groupe_votes",
	"votes",
	"acteur_vote_stats",
	"groupe_vote_stats",
}

func (s *Store) Validate(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sqlite database: %w", err)
	}

	for _, table := range requiredTables {
		var found int
		if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM sqlite_schema
WHERE type = 'table' AND name = ?
`, table).Scan(&found); err != nil {
			return fmt.Errorf("check required table %q: %w", table, err)
		}
		if found == 0 {
			return fmt.Errorf("missing required table %q", table)
		}
	}

	var schemaVersion string
	if err := s.db.QueryRowContext(ctx, `SELECT value FROM dataset_meta WHERE key = 'schema_version'`).Scan(&schemaVersion); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("missing dataset metadata %q", "schema_version")
		}
		return fmt.Errorf("query schema version: %w", err)
	}
	if schemaVersion != expectedSchemaVersion {
		return fmt.Errorf("unsupported database schema version %q, want %q", schemaVersion, expectedSchemaVersion)
	}

	return nil
}
