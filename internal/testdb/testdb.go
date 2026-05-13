package testdb

import (
	"database/sql"
	"path/filepath"
	"testing"

	"gouv.viz/internal/dbmigration"

	_ "modernc.org/sqlite"
)

func Open(t *testing.T, filename string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), filename))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	CreateSchema(t, db)
	return db
}

func CreateSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	if err := dbmigration.Run(db); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
}
