package main

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestBuildDatabaseImportsFixtureDataset(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "gouv-viz.sqlite")

	result, err := buildDatabaseWithOptions(filepath.Join("..", "..", "data", "fixtures", "raw"), outPath, buildOptions{
		AmendementResolver: stubAmendementResolver{ref: officialAmendementReference{
			TextNum:       "2695",
			Organe:        "AN",
			AmendementNum: "301",
			URL:           "https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/301.pdf",
		}},
	})
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
	assertScalar(t, db, `SELECT titre FROM scrutins WHERE uid = 'VTANR5L17V1'`, "Scrutin sur l'amendement n° 301 au projet de loi (n° 2630)")
	assertScalar(t, db, `SELECT linked_text_num FROM scrutins WHERE uid = 'VTANR5L17V1'`, "2630")
	assertScalar(t, db, `SELECT linked_text_kind FROM scrutins WHERE uid = 'VTANR5L17V1'`, "projet-loi")
	assertScalar(t, db, `SELECT linked_dossier_ref FROM scrutins WHERE uid = 'VTANR5L17V1'`, "DLR5L17N1")
	assertScalar(t, db, `SELECT linked_dossier_libelle FROM scrutins WHERE uid = 'VTANR5L17V1'`, "Projet fixture")
	assertScalar(t, db, `SELECT linked_amendement_num FROM scrutins WHERE uid = 'VTANR5L17V1'`, "301")
	assertScalar(t, db, `SELECT linked_amendement_text_num FROM scrutins WHERE uid = 'VTANR5L17V1'`, "2695")
	assertScalar(t, db, `SELECT linked_amendement_organe FROM scrutins WHERE uid = 'VTANR5L17V1'`, "AN")
	assertScalar(t, db, `SELECT linked_amendement_url FROM scrutins WHERE uid = 'VTANR5L17V1'`, "https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/301.pdf")
	assertScalar(t, db, `SELECT linked_reference_source FROM scrutins WHERE uid = 'VTANR5L17V1'`, "titre")
	assertScalar(t, db, `SELECT COUNT(*) FROM scrutin_search WHERE scrutin_search MATCH 'fixture'`, "1")
	assertScalar(t, db, `SELECT position FROM votes WHERE scrutin_uid = 'VTANR5L17V1' AND acteur_uid = 'PA100001'`, "pour")
	assertScalar(t, db, `SELECT par_delegation FROM votes WHERE scrutin_uid = 'VTANR5L17V1' AND acteur_uid = 'PA100001'`, "1")
	assertScalar(t, db, `SELECT total_votes FROM acteur_vote_stats WHERE acteur_uid = 'PA100001' AND legislature = 17`, "1")
	assertScalar(t, db, `SELECT COUNT(*) FROM acteur_latest_group WHERE acteur_uid = 'PA100001'`, "1")
	assertScalar(t, db, `SELECT deputies_count FROM groupe_member_stats WHERE groupe_uid = (SELECT groupe_uid FROM acteur_latest_group WHERE acteur_uid = 'PA100001')`, "1")
}

func TestValidateDatabaseRejectsForeignKeyViolations(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "broken.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := createSchema(db); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO acteur_adresses (uid, acteur_uid) VALUES ('ADDR1', 'missing')`); err != nil {
		t.Fatalf("insert broken row: %v", err)
	}

	err = validateDatabase(db)
	if err == nil || !strings.Contains(err.Error(), "foreign key violation") {
		t.Fatalf("validateDatabase() error = %v, want foreign key violation", err)
	}
}

type stubAmendementResolver struct {
	ref officialAmendementReference
}

func (r stubAmendementResolver) Resolve(int, string, string, string, string) (officialAmendementReference, error) {
	return r.ref, nil
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
