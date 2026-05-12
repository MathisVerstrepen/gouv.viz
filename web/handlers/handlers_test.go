package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"gouv.viz/internal/store"
	"gouv.viz/internal/testdb"
)

func TestParseScrutinsQueryParsesSearchSortAndPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins?q=%20budget%20&sort=closest&page=3&legislature=17&result=adopte&vote_type=SPO&organe=PO800538&date_from=2024-01-01&date_to=2024-12-31&close_votes=1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	query := parseScrutinsQuery(ctx)

	if query.Search != "budget" || query.Sort != "closest" || query.Page != 3 || query.PerPage != store.ScrutinsPerPage || query.Legislature != 17 || query.Result != "adopte" || query.VoteType != "SPO" || query.Organe != "PO800538" || query.DateFrom != "2024-01-01" || query.DateTo != "2024-12-31" || !query.CloseVotes {
		t.Fatalf("query = %+v, want parsed search, sort, page and filters", query)
	}
}

func TestParseScrutinsQueryIgnoresInvalidPageAndSort(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins?sort=unknown&page=-7", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	query := parseScrutinsQuery(ctx)

	if query.Page != 1 || query.Sort != "date_desc" || query.Legislature != 0 {
		t.Fatalf("query = %+v, want page 1, no legislature, and default sort date_desc", query)
	}
}

func TestScrutinsHandlerRendersOK(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins?q=budget&sort=closest&page=1", nil)
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

func TestScrutinsHandlerRendersHTMXPartial(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins?q=budget&sort=closest&page=1", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.Scrutins(ctx); err != nil {
		t.Fatalf("Scrutins() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="scrutins-explorer"`) {
		t.Fatalf("response body does not contain scrutins explorer partial")
	}
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Fatalf("response body contains full document, want partial")
	}
}

func TestScrutinDetailHandlerRendersLinkedReference(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/scrutins/scrutin-1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/scrutins/:uid")
	ctx.SetParamNames("uid")
	ctx.SetParamValues("scrutin-1")

	if err := server.ScrutinDetail(ctx); err != nil {
		t.Fatalf("ScrutinDetail() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Documents liés") || !strings.Contains(body, "2630") || !strings.Contains(body, "301") || !strings.Contains(body, "https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi") || !strings.Contains(body, "https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi.pdf") || !strings.Contains(body, "https://www.assemblee-nationale.fr/dyn/17/dossiers/DLR5L17N1") || !strings.Contains(body, "https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/301.pdf") {
		t.Fatalf("response body does not contain linked reference: %s", body)
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

func TestHTTPErrorHandlerRendersCustomPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	NewHTTPErrorHandler(e)(echo.NewHTTPError(http.StatusNotFound), ctx)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Erreur 404") || !strings.Contains(body, "La page demandée est introuvable.") {
		t.Fatalf("response body does not contain custom error content: %s", body)
	}
}

func newHandlerTestDB(t *testing.T, withScrutin bool) *sql.DB {
	t.Helper()

	db := testdb.Open(t, "handlers.sqlite")
	if _, err := db.Exec(`INSERT INTO organes (uid, libelle, libelle_abrege, preseance) VALUES ('ORG1', 'Commission fixture', 'CF', 1)`); err != nil {
		t.Fatalf("insert organe: %v", err)
	}
	if withScrutin {
		_, err := db.Exec(`
INSERT INTO scrutins (
  uid, numero, legislature, organe_uid, date_scrutin, code_type_vote,
  libelle_type_vote, type_majorite, sort_code, sort_libelle, titre,
  linked_text_num, linked_text_kind, linked_text_url, linked_text_pdf_url, linked_dossier_ref, linked_dossier_libelle,
  linked_amendement_num, linked_amendement_text_num, linked_amendement_organe,
  linked_amendement_url, linked_reference_source,
  demandeur_texte, objet_libelle, mode_publication_votes, nombre_votants,
  suffrages_exprimes, suffrages_requis, non_votants, pour, contre,
  abstentions, non_votants_volontaires, source_file
) VALUES ('scrutin-1', 1, 17, 'ORG1', '2024-07-18', 'SPO', 'Scrutin public ordinaire', 'simple', 'adopte', 'Adopte', 'Budget fixture', '2630', 'projet-loi', 'https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi', 'https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi.pdf', 'DLR5L17N1', 'Projet fixture', '301', '2695', 'AN', 'https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/301.pdf', 'titre', 'Gouvernement', 'Objet', 'Decompte', 10, 9, 5, 1, 6, 3, 0, 0, 'fixture.json')
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
