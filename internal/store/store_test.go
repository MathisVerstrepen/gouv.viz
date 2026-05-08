package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestHomePageHandlesEmptyDatabase(t *testing.T) {
	s := newTestStore(t)

	page, err := s.HomePage(context.Background())
	if err != nil {
		t.Fatalf("HomePage() error = %v", err)
	}

	if page.TotalScrutins != 0 {
		t.Fatalf("TotalScrutins = %d, want 0", page.TotalScrutins)
	}
	if page.FirstScrutinDate != "" || page.LastScrutinDate != "" {
		t.Fatalf("date range = %q/%q, want empty strings", page.FirstScrutinDate, page.LastScrutinDate)
	}
	if len(page.Scrutins) != 0 {
		t.Fatalf("len(Scrutins) = %d, want 0", len(page.Scrutins))
	}
}

func TestScrutinsPageEscapesLikeWildcards(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertScrutin(t, s.db, testScrutin{UID: "plain", Numero: 1, Date: "2024-01-01", Titre: "Budget ordinaire", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "percent", Numero: 2, Date: "2024-01-02", Titre: "Budget 100% public", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "underscore", Numero: 3, Date: "2024-01-03", Titre: "Article A_B", OrganeUID: "ORG1"})

	percentPage, err := s.ScrutinsPage(context.Background(), ScrutinsQuery{Search: "%", Page: 1, PerPage: 25})
	if err != nil {
		t.Fatalf("ScrutinsPage(%%) error = %v", err)
	}
	if got := uids(percentPage.Scrutins); len(got) != 1 || got[0] != "percent" {
		t.Fatalf("search %% returned %v, want [percent]", got)
	}

	underscorePage, err := s.ScrutinsPage(context.Background(), ScrutinsQuery{Search: "_", Page: 1, PerPage: 25})
	if err != nil {
		t.Fatalf("ScrutinsPage(_) error = %v", err)
	}
	if got := uids(underscorePage.Scrutins); len(got) != 1 || got[0] != "underscore" {
		t.Fatalf("search _ returned %v, want [underscore]", got)
	}
}

func TestScrutinsPageClampsPagePastTotal(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertScrutin(t, s.db, testScrutin{UID: "one", Numero: 1, Date: "2024-01-01", Titre: "Premier", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "two", Numero: 2, Date: "2024-01-02", Titre: "Deuxieme", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "three", Numero: 3, Date: "2024-01-03", Titre: "Troisieme", OrganeUID: "ORG1"})

	page, err := s.ScrutinsPage(context.Background(), ScrutinsQuery{Sort: "numero_asc", Page: 99, PerPage: 2})
	if err != nil {
		t.Fatalf("ScrutinsPage() error = %v", err)
	}

	if page.Query.Page != 2 || page.TotalPages != 2 || page.StartItem != 3 || page.EndItem != 3 {
		t.Fatalf("pagination = page %d total %d start %d end %d, want page 2 total 2 start 3 end 3", page.Query.Page, page.TotalPages, page.StartItem, page.EndItem)
	}
	if got := uids(page.Scrutins); len(got) != 1 || got[0] != "three" {
		t.Fatalf("scrutins = %v, want [three]", got)
	}
}

func TestScrutinDetailPageReturnsErrNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.ScrutinDetailPage(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("ScrutinDetailPage() error = %v, want ErrNotFound", err)
	}
}

func TestScrutinDetailPageOrdersGroupVotesByPreseance(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "GRP2", "Groupe deux", "G2", 2)
	insertOrgane(t, s.db, "GRP1", "Groupe un", "G1", 1)
	insertScrutin(t, s.db, testScrutin{UID: "detail", Numero: 7, Date: "2024-02-01", Titre: "Vote detaille", OrganeUID: "GRP1"})
	insertGroupVote(t, s.db, "detail", "GRP2", "contre")
	insertGroupVote(t, s.db, "detail", "GRP1", "pour")

	page, err := s.ScrutinDetailPage(context.Background(), "detail")
	if err != nil {
		t.Fatalf("ScrutinDetailPage() error = %v", err)
	}

	if len(page.GroupVotes) != 2 {
		t.Fatalf("len(GroupVotes) = %d, want 2", len(page.GroupVotes))
	}
	if page.GroupVotes[0].GroupeUID != "GRP1" || page.GroupVotes[1].GroupeUID != "GRP2" {
		t.Fatalf("group order = %s, %s; want GRP1, GRP2", page.GroupVotes[0].GroupeUID, page.GroupVotes[1].GroupeUID)
	}
}

func TestValidateAcceptsExpectedDatabase(t *testing.T) {
	s := newValidationTestStore(t, expectedSchemaVersion, nil)

	if err := s.Validate(context.Background()); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsMissingRequiredTable(t *testing.T) {
	s := newValidationTestStore(t, expectedSchemaVersion, map[string]bool{"votes": true})

	err := s.Validate(context.Background())
	if err == nil || !strings.Contains(err.Error(), `missing required table "votes"`) {
		t.Fatalf("Validate() error = %v, want missing votes table", err)
	}
}

func TestValidateRejectsUnexpectedSchemaVersion(t *testing.T) {
	s := newValidationTestStore(t, "0", nil)

	err := s.Validate(context.Background())
	if err == nil || !strings.Contains(err.Error(), `unsupported database schema version "0"`) {
		t.Fatalf("Validate() error = %v, want unsupported schema version", err)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(testSchema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	return New(db)
}

func newValidationTestStore(t *testing.T, schemaVersion string, skipTables map[string]bool) *Store {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "validation.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	for _, table := range requiredTables {
		if skipTables[table] {
			continue
		}

		statement := "CREATE TABLE " + table + " (id INTEGER)"
		if table == "dataset_meta" {
			statement = "CREATE TABLE dataset_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL)"
		}
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("create table %s: %v", table, err)
		}
	}

	if !skipTables["dataset_meta"] {
		if _, err := db.Exec(`INSERT INTO dataset_meta (key, value) VALUES ('schema_version', ?)`, schemaVersion); err != nil {
			t.Fatalf("insert schema version: %v", err)
		}
	}

	return New(db)
}

const testSchema = `
CREATE TABLE organes (
  uid TEXT PRIMARY KEY,
  libelle TEXT,
  libelle_abrege TEXT,
  preseance INTEGER
);

CREATE TABLE scrutins (
  uid TEXT PRIMARY KEY,
  numero INTEGER NOT NULL,
  legislature INTEGER NOT NULL,
  organe_uid TEXT,
  session_ref TEXT,
  seance_ref TEXT,
  date_scrutin TEXT NOT NULL,
  quantieme_jour_seance INTEGER,
  code_type_vote TEXT,
  libelle_type_vote TEXT,
  type_majorite TEXT,
  sort_code TEXT,
  sort_libelle TEXT,
  titre TEXT,
  demandeur_texte TEXT,
  objet_libelle TEXT,
  mode_publication_votes TEXT,
  nombre_votants INTEGER,
  suffrages_exprimes INTEGER,
  suffrages_requis INTEGER,
  non_votants INTEGER,
  pour INTEGER,
  contre INTEGER,
  abstentions INTEGER,
  non_votants_volontaires INTEGER,
  source_file TEXT
);

CREATE TABLE scrutin_groupe_votes (
  scrutin_uid TEXT NOT NULL,
  groupe_uid TEXT NOT NULL,
  nombre_membres_groupe INTEGER,
  position_majoritaire TEXT,
  non_votants INTEGER,
  pour INTEGER,
  contre INTEGER,
  abstentions INTEGER,
  non_votants_volontaires INTEGER,
  PRIMARY KEY (scrutin_uid, groupe_uid)
);
`

type testScrutin struct {
	UID       string
	Numero    int
	Date      string
	Titre     string
	OrganeUID string
}

func insertOrgane(t *testing.T, db *sql.DB, uid, libelle, libelleAbrege string, preseance int) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO organes (uid, libelle, libelle_abrege, preseance) VALUES (?, ?, ?, ?)`, uid, libelle, libelleAbrege, preseance)
	if err != nil {
		t.Fatalf("insert organe %s: %v", uid, err)
	}
}

func insertScrutin(t *testing.T, db *sql.DB, scrutin testScrutin) {
	t.Helper()
	_, err := db.Exec(`
INSERT INTO scrutins (
  uid, numero, legislature, organe_uid, date_scrutin, code_type_vote,
  libelle_type_vote, type_majorite, sort_code, sort_libelle, titre,
  demandeur_texte, objet_libelle, mode_publication_votes, nombre_votants,
  suffrages_exprimes, suffrages_requis, non_votants, pour, contre,
  abstentions, non_votants_volontaires, source_file
) VALUES (?, ?, 17, ?, ?, 'SPO', 'Scrutin public ordinaire', 'simple', 'adopte', 'Adopte', ?, 'Gouvernement', 'Objet', 'Decompte', 10, 9, 5, 1, 6, 3, 0, 0, 'fixture.json')
`, scrutin.UID, scrutin.Numero, scrutin.OrganeUID, scrutin.Date, scrutin.Titre)
	if err != nil {
		t.Fatalf("insert scrutin %s: %v", scrutin.UID, err)
	}
}

func insertGroupVote(t *testing.T, db *sql.DB, scrutinUID, groupeUID, position string) {
	t.Helper()
	_, err := db.Exec(`
INSERT INTO scrutin_groupe_votes (scrutin_uid, groupe_uid, nombre_membres_groupe, position_majoritaire, non_votants, pour, contre, abstentions, non_votants_volontaires)
VALUES (?, ?, 10, ?, 1, 6, 3, 0, 0)
`, scrutinUID, groupeUID, position)
	if err != nil {
		t.Fatalf("insert group vote %s/%s: %v", scrutinUID, groupeUID, err)
	}
}

func uids(items []ScrutinListItem) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.UID)
	}
	return values
}
