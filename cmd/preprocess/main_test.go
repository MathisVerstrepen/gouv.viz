package main

import (
	"database/sql"
	"os"
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

func TestBuildDatabaseReportsUnmatchedAmendement(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "gouv-viz.sqlite")

	result, err := buildDatabaseWithOptions(filepath.Join("..", "..", "data", "fixtures", "raw"), outPath, buildOptions{
		AmendementResolver: stubAmendementResolver{},
	})
	if err != nil {
		t.Fatalf("buildDatabase() error = %v", err)
	}

	if got := result.Diagnostics.Count("unmatched_amendement"); got != 1 {
		t.Fatalf("unmatched_amendement diagnostics = %d, want 1", got)
	}
	if report := result.Diagnostics.Report(); !strings.Contains(report, "amendment 301") {
		t.Fatalf("diagnostics report = %q, want amendment example", report)
	}
}

func TestBuildDatabaseReportsUnresolvedReferences(t *testing.T) {
	rawDir := writeMinimalRawDataset(t, `{
  "scrutin": {
    "uid": "VT1",
    "numero": "1",
    "legislature": "17",
    "organeRef": "PO_MISSING",
    "dateScrutin": "2024-07-18",
    "ventilationVotes": {
      "organe": {
        "groupes": {
          "groupe": {
            "organeRef": "PO_MISSING",
            "vote": {
              "positionMajoritaire": "mystere",
              "decompteNominatif": {
                "pours": {"votant": {"acteurRef": "PA1"}}
              }
            }
          }
        }
      }
    }
  }
}`)
	outPath := filepath.Join(t.TempDir(), "gouv-viz.sqlite")

	result, err := buildDatabase(rawDir, outPath)
	if err == nil {
		t.Fatalf("buildDatabase() error = nil, want validation error")
	}
	if got := result.Diagnostics.Count("unresolved_organe"); got == 0 {
		t.Fatalf("unresolved_organe diagnostics = %d, want at least 1", got)
	}
	if got := result.Diagnostics.Count("unknown_vote_position"); got == 0 {
		t.Fatalf("unknown_vote_position diagnostics = %d, want at least 1", got)
	}
	message := err.Error()
	if !strings.Contains(message, "foreign key violation") || !strings.Contains(message, "import anomalies") || !strings.Contains(message, "unresolved_organe") {
		t.Fatalf("buildDatabase() error = %q, want validation error with anomaly report", message)
	}
}

func TestBuildDatabaseReportsMissingScrutinDate(t *testing.T) {
	rawDir := writeMinimalRawDataset(t, `{
  "scrutin": {
    "uid": "VT1",
    "numero": "1",
    "legislature": "17"
  }
}`)
	outPath := filepath.Join(t.TempDir(), "gouv-viz.sqlite")

	result, err := buildDatabase(rawDir, outPath)
	if err == nil || !strings.Contains(err.Error(), "missing dateScrutin") {
		t.Fatalf("buildDatabase() error = %v, want missing dateScrutin", err)
	}
	if got := result.Diagnostics.Count("missing_date"); got != 1 {
		t.Fatalf("missing_date diagnostics = %d, want 1", got)
	}
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

func writeMinimalRawDataset(t *testing.T, scrutinJSON string) string {
	t.Helper()

	rawDir := t.TempDir()
	files := map[string]string{
		filepath.Join(rawDir, "organe", "PO800001.json"):      `{"organe":{"uid":"PO800001","codeType":"GP","libelle":"Groupe fixture"}}`,
		filepath.Join(rawDir, "acteur", "PA1.json"):           `{"acteur":{"uid":"PA1"}}`,
		filepath.Join(rawDir, "scrutins-publics", "VT1.json"): scrutinJSON,
	}
	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create fixture directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write fixture %s: %v", path, err)
		}
	}
	return rawDir
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
