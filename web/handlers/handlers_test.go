package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"

	"gouv.viz/internal/store"
)

func TestParseScrutinsQueryParsesSearchSortAndPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins?q=%20budget%20&sort=numero_asc&page=3", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	query := parseScrutinsQuery(ctx)

	if query.Search != "budget" || query.Sort != "numero_asc" || query.Page != 3 || query.PerPage != store.ScrutinsPerPage {
		t.Fatalf("query = %+v, want search budget, sort numero_asc, page 3, per-page %d", query, store.ScrutinsPerPage)
	}
}

func TestParseScrutinsQueryIgnoresInvalidPageAndSort(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins?sort=unknown&page=-7", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	query := parseScrutinsQuery(ctx)

	if query.Page != 1 || query.Sort != "date_desc" {
		t.Fatalf("query = %+v, want page 1 and default sort date_desc", query)
	}
}

func TestScrutinsHandlerRendersOK(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins?q=budget&sort=numero_asc&page=1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.Scrutins(ctx); err != nil {
		t.Fatalf("Scrutins() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Budget fixture") {
		t.Fatalf("response body does not contain inserted scrutin title")
	}
}

func TestScrutinDetailHandlerReturns404ForMissing(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, false)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins/missing", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/scrutins/:uid")
	ctx.SetParamNames("uid")
	ctx.SetParamValues("missing")

	err := server.ScrutinDetail(ctx)
	if err == nil {
		t.Fatal("ScrutinDetail() error = nil, want HTTP 404")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusNotFound {
		t.Fatalf("ScrutinDetail() error = %#v, want HTTP 404", err)
	}
}

func newHandlerTestDB(t *testing.T, withScrutin bool) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "handlers.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(handlerTestSchema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO organes (uid, libelle, libelle_abrege, preseance) VALUES ('ORG1', 'Commission fixture', 'CF', 1)`); err != nil {
		t.Fatalf("insert organe: %v", err)
	}
	if withScrutin {
		_, err := db.Exec(`
INSERT INTO scrutins (
  uid, numero, legislature, organe_uid, date_scrutin, code_type_vote,
  libelle_type_vote, type_majorite, sort_code, sort_libelle, titre,
  demandeur_texte, objet_libelle, mode_publication_votes, nombre_votants,
  suffrages_exprimes, suffrages_requis, non_votants, pour, contre,
  abstentions, non_votants_volontaires, source_file
) VALUES ('scrutin-1', 1, 17, 'ORG1', '2024-07-18', 'SPO', 'Scrutin public ordinaire', 'simple', 'adopte', 'Adopte', 'Budget fixture', 'Gouvernement', 'Objet', 'Decompte', 10, 9, 5, 1, 6, 3, 0, 0, 'fixture.json')
`)
		if err != nil {
			t.Fatalf("insert scrutin: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO scrutin_search (uid, document) VALUES ('scrutin-1', 'Budget fixture')`); err != nil {
			t.Fatalf("insert scrutin search: %v", err)
		}
	}

	return db
}

const handlerTestSchema = `
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

CREATE VIRTUAL TABLE scrutin_search USING fts5(
  uid UNINDEXED,
  document,
  tokenize = 'unicode61 remove_diacritics 2'
);
`
