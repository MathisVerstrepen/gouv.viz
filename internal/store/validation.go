package store

import (
	"context"
	"database/sql"
	"fmt"

	"gouv.viz/internal/dbmigration"
)

const expectedSchemaVersion = dbmigration.LatestVersionString

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
	"acteur_latest_group",
	"groupe_member_stats",
	"scrutin_search",
	"schema_migrations",
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
	var dirty int
	if err := s.db.QueryRowContext(ctx, `SELECT CAST(version AS TEXT), dirty FROM schema_migrations`).Scan(&schemaVersion, &dirty); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("missing schema migration version")
		}
		return fmt.Errorf("query schema version: %w", err)
	}
	if dirty != 0 {
		return fmt.Errorf("dirty database schema migration version %q", schemaVersion)
	}
	if schemaVersion != expectedSchemaVersion {
		return fmt.Errorf("unsupported database schema version %q, want %q", schemaVersion, expectedSchemaVersion)
	}

	return nil
}
