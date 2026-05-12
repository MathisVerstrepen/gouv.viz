package testdb

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

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

	if _, err := db.Exec(SchemaSQL(t)); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
}

func SchemaSQL(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve testdb source path")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "cmd", "preprocess", "schema.sql")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema.sql: %v", err)
	}
	return string(contents)
}
