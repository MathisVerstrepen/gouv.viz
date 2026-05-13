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

func TestParseDeputiesQueryParsesSearchSortAndFilters(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/deputes?q=%20alice%20&sort=votes_desc&page=2&legislature=17&group=GRP1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	query := parseDeputiesQuery(ctx)
	if query.Search != "alice" || query.Sort != "votes_desc" || query.Page != 2 || query.PerPage != store.DeputiesPerPage || query.Legislature != 17 || query.Group != "GRP1" {
		t.Fatalf("query = %+v, want parsed deputies filters", query)
	}
}

func TestDeputiesHandlerRendersOK(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/deputes?q=alice", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.Deputies(ctx); err != nil {
		t.Fatalf("Deputies() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Alice Martin") || !strings.Contains(rec.Body.String(), "Tous les députés") {
		t.Fatalf("response body does not contain deputies page content: %s", rec.Body.String())
	}
}

func TestDeputiesHandlerRendersHTMXPartial(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/deputes?q=alice", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.Deputies(ctx); err != nil {
		t.Fatalf("Deputies() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="deputies-explorer"`) || !strings.Contains(body, "Alice Martin") {
		t.Fatalf("response body does not contain deputies explorer partial: %s", body)
	}
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Fatalf("response body contains full document, want partial")
	}
}

func TestParsePoliticalGroupsQueryParsesSearchSortAndFilters(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/groupes?q=%20gauche%20&sort=deputies_desc&page=2&legislature=17", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	query := parsePoliticalGroupsQuery(ctx)
	if query.Search != "gauche" || query.Sort != "deputies_desc" || query.Page != 2 || query.PerPage != store.PoliticalGroupsPerPage || query.Legislature != 17 {
		t.Fatalf("query = %+v, want parsed political group filters", query)
	}
}

func TestPoliticalGroupsHandlerRendersOK(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/groupes?q=fixture", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.PoliticalGroups(ctx); err != nil {
		t.Fatalf("PoliticalGroups() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Groupe fixture") || !strings.Contains(rec.Body.String(), "Forces politiques") {
		t.Fatalf("response body does not contain political groups page content: %s", rec.Body.String())
	}
}

func TestPoliticalGroupsHandlerRendersHTMXPartial(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/groupes?q=fixture", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.PoliticalGroups(ctx); err != nil {
		t.Fatalf("PoliticalGroups() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="political-groups-explorer"`) || !strings.Contains(body, "Groupe fixture") {
		t.Fatalf("response body does not contain political groups explorer partial: %s", body)
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

func TestDeputyDetailHandlerRendersOK(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/deputes/PA1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/deputes/:uid")
	ctx.SetParamNames("uid")
	ctx.SetParamValues("PA1")

	if err := server.DeputyDetail(ctx); err != nil {
		t.Fatalf("DeputyDetail() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Alice Martin") || !strings.Contains(body, "Votes nominatifs") || !strings.Contains(body, "Budget fixture") {
		t.Fatalf("response body does not contain deputy detail content: %s", body)
	}
}

func TestDeputyDetailHandlerReturns404ForMissing(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, false)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/deputes/missing", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/deputes/:uid")
	ctx.SetParamNames("uid")
	ctx.SetParamValues("missing")

	err := server.DeputyDetail(ctx)
	if err == nil {
		t.Fatal("DeputyDetail() error = nil, want HTTP 404")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusNotFound {
		t.Fatalf("DeputyDetail() error = %#v, want HTTP 404", err)
	}
}

func TestDeputyDetailHandlerRendersVotesPanelForHTMX(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/deputes/PA1?votes_page=1", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/deputes/:uid")
	ctx.SetParamNames("uid")
	ctx.SetParamValues("PA1")

	if err := server.DeputyDetail(ctx); err != nil {
		t.Fatalf("DeputyDetail() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="depute-votes-panel"`) || !strings.Contains(body, "Budget fixture") {
		t.Fatalf("response body does not contain votes panel content: %s", body)
	}
	if strings.Contains(body, "<!doctype html") || strings.Contains(body, "SiteHeader") {
		t.Fatalf("response body contains full-page content: %s", body)
	}
}

func TestParseDeputyDetailQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/deputes/PA1?votes_page=3&votes_q=budget&votes_sort=date_asc&votes_position=pour", nil)
	ctx := e.NewContext(req, httptest.NewRecorder())

	query := parseDeputyDetailQuery(ctx)
	if query.VotesPage != 3 || query.VotesPerPage != store.DeputyVotesPerPage || query.VotesSearch != "budget" || query.VotesSort != "date_asc" || query.VotesPosition != "pour" {
		t.Fatalf("parseDeputyDetailQuery() = %+v", query)
	}
}

func TestParsePoliticalGroupDetailQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/groupes/GRP1?votes_page=3&votes_q=budget&votes_sort=date_asc&votes_position=pour", nil)
	ctx := e.NewContext(req, httptest.NewRecorder())

	query := parsePoliticalGroupDetailQuery(ctx)
	if query.VotesPage != 3 || query.VotesPerPage != store.PoliticalGroupVotesPerPage || query.VotesSearch != "budget" || query.VotesSort != "date_asc" || query.VotesPosition != "pour" {
		t.Fatalf("parsePoliticalGroupDetailQuery() = %+v", query)
	}
}

func TestPoliticalGroupDetailHandlerRendersOK(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/groupes/GRP1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/groupes/:uid")
	ctx.SetParamNames("uid")
	ctx.SetParamValues("GRP1")

	if err := server.PoliticalGroupDetail(ctx); err != nil {
		t.Fatalf("PoliticalGroupDetail() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Groupe fixture") || !strings.Contains(body, "Historique des votes") || !strings.Contains(body, "Budget fixture") {
		t.Fatalf("response body does not contain political group detail content: %s", body)
	}
}

func TestPoliticalGroupDetailHandlerRendersVotesPanelForHTMX(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/groupes/GRP1?votes_page=1", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/groupes/:uid")
	ctx.SetParamNames("uid")
	ctx.SetParamValues("GRP1")

	if err := server.PoliticalGroupDetail(ctx); err != nil {
		t.Fatalf("PoliticalGroupDetail() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="groupe-votes-panel"`) || !strings.Contains(body, "Budget fixture") {
		t.Fatalf("response body does not contain group votes panel content: %s", body)
	}
	if strings.Contains(body, "<!DOCTYPE html>") || strings.Contains(body, "Groupe politique") {
		t.Fatalf("response body contains full-page content: %s", body)
	}
}

func TestPoliticalGroupDetailHandlerReturns404ForMissing(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, false)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/groupes/missing", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/groupes/:uid")
	ctx.SetParamNames("uid")
	ctx.SetParamValues("missing")

	err := server.PoliticalGroupDetail(ctx)
	if err == nil {
		t.Fatal("PoliticalGroupDetail() error = nil, want HTTP 404")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusNotFound {
		t.Fatalf("PoliticalGroupDetail() error = %#v, want HTTP 404", err)
	}
}

func TestSitemapIndexReturnsXML(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.SitemapIndex(SitemapConfig{BaseURL: "https://example.com"})(ctx); err != nil {
		t.Fatalf("SitemapIndex() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/xml; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/xml; charset=utf-8", contentType)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Fatalf("response missing XML declaration")
	}
	if !strings.Contains(body, `<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`) {
		t.Fatalf("response missing sitemapindex")
	}
	if !strings.Contains(body, "https://example.com/sitemap-static.xml") {
		t.Fatalf("response missing static sitemap ref")
	}
	if !strings.Contains(body, "https://example.com/sitemap-scrutins.xml") {
		t.Fatalf("response missing scrutins sitemap ref")
	}
	if !strings.Contains(body, "https://example.com/sitemap-deputes.xml") {
		t.Fatalf("response missing deputes sitemap ref")
	}
	if !strings.Contains(body, "https://example.com/sitemap-groupes.xml") {
		t.Fatalf("response missing groupes sitemap ref")
	}
}

func TestSitemapStaticReturnsXML(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/sitemap-static.xml", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.SitemapStatic(SitemapConfig{BaseURL: "https://example.com"})(ctx); err != nil {
		t.Fatalf("SitemapStatic() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "https://example.com/") {
		t.Fatalf("response missing home URL")
	}
	if !strings.Contains(body, "https://example.com/scrutins") {
		t.Fatalf("response missing scrutins list URL")
	}
	if !strings.Contains(body, "https://example.com/deputes") {
		t.Fatalf("response missing deputes list URL")
	}
	if !strings.Contains(body, "https://example.com/groupes") {
		t.Fatalf("response missing groupes list URL")
	}
	if strings.Contains(body, "scrutin-1") {
		t.Fatalf("static sitemap should not contain detail URLs")
	}
}

func TestSitemapScrutinsReturnsXML(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/sitemap-scrutins.xml", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.SitemapScrutins(SitemapConfig{BaseURL: "https://example.com"})(ctx); err != nil {
		t.Fatalf("SitemapScrutins() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "https://example.com/scrutins/scrutin-1") {
		t.Fatalf("response missing scrutin detail URL")
	}
}

func TestSitemapDeputiesReturnsXML(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/sitemap-deputes.xml", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.SitemapDeputies(SitemapConfig{BaseURL: "https://example.com"})(ctx); err != nil {
		t.Fatalf("SitemapDeputies() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "https://example.com/deputes/PA1") {
		t.Fatalf("response missing deputy detail URL")
	}
}

func TestSitemapGroupsReturnsXML(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/sitemap-groupes.xml", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.SitemapGroups(SitemapConfig{BaseURL: "https://example.com"})(ctx); err != nil {
		t.Fatalf("SitemapGroups() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "https://example.com/groupes/GRP1") {
		t.Fatalf("response missing group detail URL")
	}
}

func TestSitemapIndexDerivesBaseURLFromRequest(t *testing.T) {
	server := NewServer(store.New(newHandlerTestDB(t, true)))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	req.Host = "gouv.viz.local"
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.SitemapIndex(SitemapConfig{})(ctx); err != nil {
		t.Fatalf("SitemapIndex() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "http://gouv.viz.local/sitemap-static.xml") {
		t.Fatalf("response missing derived base URL")
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
	if _, err := db.Exec(`INSERT INTO organes (uid, code_type, libelle, libelle_abrege, libelle_abrev, legislature, position_politique, date_debut, preseance) VALUES ('GRP1', 'GP', 'Groupe fixture', 'GF', 'GF', 17, 'Centre', '2024-01-01', 2)`); err != nil {
		t.Fatalf("insert group organe: %v", err)
	}
	if withScrutin {
		if _, err := db.Exec(`INSERT INTO acteurs (uid, prenom, nom, alpha, profession, source_file) VALUES ('PA1', 'Alice', 'Martin', 'MARTIN', 'Juriste', 'acteur.json')`); err != nil {
			t.Fatalf("insert acteur: %v", err)
		}
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
		if _, err := db.Exec(`INSERT INTO votes (scrutin_uid, acteur_uid, groupe_uid, position, par_delegation, num_place) VALUES ('scrutin-1', 'PA1', 'GRP1', 'pour', 0, '12')`); err != nil {
			t.Fatalf("insert vote: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO scrutin_groupe_votes (scrutin_uid, groupe_uid, nombre_membres_groupe, position_majoritaire, non_votants, pour, contre, abstentions, non_votants_volontaires) VALUES ('scrutin-1', 'GRP1', 10, 'pour', 1, 6, 3, 0, 0)`); err != nil {
			t.Fatalf("insert scrutin group vote: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO groupe_vote_stats (groupe_uid, legislature, total_scrutins, pour, contre, abstentions, non_votants) VALUES ('GRP1', 17, 1, 1, 0, 0, 0)`); err != nil {
			t.Fatalf("insert group vote stats: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO acteur_vote_stats (acteur_uid, legislature, total_votes, pour, contre, abstentions, non_votants) VALUES ('PA1', 17, 1, 1, 0, 0, 0)`); err != nil {
			t.Fatalf("insert acteur vote stats: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO mandats (uid, acteur_uid, legislature, type_organe, date_debut, nomin_principale, lib_qualite) VALUES ('M1', 'PA1', 17, 'ASSEMBLEE', '2024-01-01', 1, 'Députée')`); err != nil {
			t.Fatalf("insert mandat: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO mandat_organes (mandat_uid, organe_uid) VALUES ('M1', 'GRP1')`); err != nil {
			t.Fatalf("insert mandat organe: %v", err)
		}
	}

	return db
}
