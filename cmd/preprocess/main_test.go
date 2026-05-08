package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestBuildDatabaseImportsFixtureDataset(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "gouv-viz.sqlite")

	result, err := buildDatabase(filepath.Join("..", "..", "data", "fixtures", "raw"), outPath)
	if err != nil {
		t.Fatalf("buildDatabase() error = %v", err)
	}

	if result.Organes != 2 {
		t.Fatalf("Organes = %d, want fixture organe plus synthetic PO0", result.Organes)
	}
	if result.Acteurs != 1 || result.Adresses != 1 || result.Mandats != 1 || result.MandatOrganes != 1 {
		t.Fatalf("actor import stats = acteurs:%d adresses:%d mandats:%d mandat_organes:%d, want 1 each", result.Acteurs, result.Adresses, result.Mandats, result.MandatOrganes)
	}
	if result.Scrutins != 1 || result.ScrutinGroupeVotes != 1 || result.Votes != 1 {
		t.Fatalf("scrutin import stats = scrutins:%d group_votes:%d votes:%d, want 1 each", result.Scrutins, result.ScrutinGroupeVotes, result.Votes)
	}

	db, err := sql.Open("sqlite", outPath)
	if err != nil {
		t.Fatalf("open output database: %v", err)
	}
	defer db.Close()

	assertScalar(t, db, `SELECT value FROM dataset_meta WHERE key = 'schema_version'`, schemaVersion)
	assertScalar(t, db, `SELECT libelle_abrege FROM organes WHERE uid = 'PO0'`, "NI")
	assertScalar(t, db, `SELECT titre FROM scrutins WHERE uid = 'VTANR5L17V1'`, "Projet fixture 100%_public")
	assertScalar(t, db, `SELECT position FROM votes WHERE scrutin_uid = 'VTANR5L17V1' AND acteur_uid = 'PA100001'`, "pour")
	assertScalar(t, db, `SELECT par_delegation FROM votes WHERE scrutin_uid = 'VTANR5L17V1' AND acteur_uid = 'PA100001'`, "1")
	assertScalar(t, db, `SELECT total_votes FROM acteur_vote_stats WHERE acteur_uid = 'PA100001' AND legislature = 17`, "1")
}

func assertScalar(t *testing.T, db *sql.DB, query string, want string) {
	t.Helper()

	var got string
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("query %q = %q, want %q", query, got, want)
	}
}
