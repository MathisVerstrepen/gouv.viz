package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"gouv.viz/internal/testdb"
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

func TestScrutinsPageSearchesFTSTokens(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertScrutin(t, s.db, testScrutin{UID: "plain", Numero: 1, Date: "2024-01-01", Titre: "Budget ordinaire", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "percent", Numero: 2, Date: "2024-01-02", Titre: "Budget 100% public", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "underscore", Numero: 3, Date: "2024-01-03", Titre: "Article A_B", OrganeUID: "ORG1"})

	percentPage, err := s.ScrutinsPage(context.Background(), ScrutinsQuery{Search: "100%", Page: 1, PerPage: 25})
	if err != nil {
		t.Fatalf("ScrutinsPage(100%%) error = %v", err)
	}
	if got := uids(percentPage.Scrutins); len(got) != 1 || got[0] != "percent" {
		t.Fatalf("search 100%% returned %v, want [percent]", got)
	}

	underscorePage, err := s.ScrutinsPage(context.Background(), ScrutinsQuery{Search: "A_B", Page: 1, PerPage: 25})
	if err != nil {
		t.Fatalf("ScrutinsPage(A_B) error = %v", err)
	}
	if got := uids(underscorePage.Scrutins); len(got) != 1 || got[0] != "underscore" {
		t.Fatalf("search A_B returned %v, want [underscore]", got)
	}
}

func TestScrutinsPageClampsPagePastTotal(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertScrutin(t, s.db, testScrutin{UID: "one", Numero: 1, Date: "2024-01-01", Titre: "Premier", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "two", Numero: 2, Date: "2024-01-02", Titre: "Deuxieme", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "three", Numero: 3, Date: "2024-01-03", Titre: "Troisieme", OrganeUID: "ORG1"})

	page, err := s.ScrutinsPage(context.Background(), ScrutinsQuery{Sort: "date_asc", Page: 99, PerPage: 2})
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

func TestScrutinsPageSortsClosestPourContreFirst(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertScrutin(t, s.db, testScrutin{UID: "wide", Numero: 1, Date: "2024-01-01", Titre: "Ecart large", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "tie", Numero: 2, Date: "2024-01-02", Titre: "Egalite", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "close", Numero: 3, Date: "2024-01-03", Titre: "Ecart serre", OrganeUID: "ORG1"})
	updateScrutinVoteCounts(t, s.db, "wide", 90, 10)
	updateScrutinVoteCounts(t, s.db, "tie", 50, 50)
	updateScrutinVoteCounts(t, s.db, "close", 49, 51)

	page, err := s.ScrutinsPage(context.Background(), ScrutinsQuery{Sort: "closest", Page: 1, PerPage: 25})
	if err != nil {
		t.Fatalf("ScrutinsPage() error = %v", err)
	}

	if got := uids(page.Scrutins); len(got) != 3 || got[0] != "tie" || got[1] != "close" || got[2] != "wide" {
		t.Fatalf("scrutins = %v, want [tie close wide]", got)
	}
}

func TestScrutinsPageAppliesStructuredFilters(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertOrgane(t, s.db, "GRP1", "Groupe un", "G1", 2)
	insertOrgane(t, s.db, "ORG2", "Autre organe", "AO", 3)
	insertScrutin(t, s.db, testScrutin{UID: "match", Numero: 1, Legislature: 17, Date: "2024-05-10", Titre: "Texte retenu", OrganeUID: "ORG1", Result: "adopte", ResultLabel: "Adopté", VoteType: "SPO", VoteTypeLabel: "Scrutin public ordinaire", Pour: 54, Contre: 49})
	insertScrutin(t, s.db, testScrutin{UID: "old", Numero: 2, Legislature: 17, Date: "2024-04-01", Titre: "Date exclue", OrganeUID: "ORG1", Result: "adopte", VoteType: "SPO", Pour: 54, Contre: 49})
	insertScrutin(t, s.db, testScrutin{UID: "wide", Numero: 3, Legislature: 17, Date: "2024-05-11", Titre: "Large exclu", OrganeUID: "ORG1", Result: "adopte", VoteType: "SPO", Pour: 90, Contre: 10})
	insertScrutin(t, s.db, testScrutin{UID: "other", Numero: 4, Legislature: 16, Date: "2024-05-10", Titre: "Autre exclu", OrganeUID: "ORG2", Result: "rejete", VoteType: "MOC", Pour: 49, Contre: 54})
	insertGroupVote(t, s.db, "match", "GRP1", "pour")

	page, err := s.ScrutinsPage(context.Background(), ScrutinsQuery{
		Legislature: 17,
		Result:      "adopte",
		VoteType:    "SPO",
		Organe:      "GRP1",
		DateFrom:    "2024-05-01",
		DateTo:      "2024-05-31",
		CloseVotes:  true,
		Page:        1,
		PerPage:     25,
	})
	if err != nil {
		t.Fatalf("ScrutinsPage() error = %v", err)
	}

	if got := uids(page.Scrutins); len(got) != 1 || got[0] != "match" {
		t.Fatalf("scrutins = %v, want [match]", got)
	}
	if len(page.FilterOptions.Legislatures) == 0 || page.FilterOptions.Legislatures[0].Value != "17" {
		t.Fatalf("legislature filter options = %+v, want 17 first", page.FilterOptions.Legislatures)
	}
	if !hasFilterOption(page.FilterOptions.Organes, "GRP1") {
		t.Fatalf("organe filter options = %+v, want GRP1 from group votes", page.FilterOptions.Organes)
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

func TestScrutinDetailPageReturnsIndividualVotesByGroup(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "GRP2", "Groupe deux", "G2", 2)
	insertOrgane(t, s.db, "GRP1", "Groupe un", "G1", 1)
	insertScrutin(t, s.db, testScrutin{UID: "detail", Numero: 7, Date: "2024-02-01", Titre: "Vote detaille", OrganeUID: "GRP1"})
	insertActeur(t, s.db, "PA1", "Alice", "Martin", "MARTIN")
	insertActeur(t, s.db, "PA2", "Bruno", "Bernard", "BERNARD")
	insertActeur(t, s.db, "PA3", "Claire", "Durand", "DURAND")
	insertIndividualVote(t, s.db, "detail", "PA2", "GRP1", "contre", false, "")
	insertIndividualVote(t, s.db, "detail", "PA1", "GRP1", "pour", true, "12")
	insertIndividualVote(t, s.db, "detail", "PA3", "GRP2", "abstention", false, "")

	page, err := s.ScrutinDetailPage(context.Background(), "detail")
	if err != nil {
		t.Fatalf("ScrutinDetailPage() error = %v", err)
	}

	if len(page.IndividualVotes) != 3 {
		t.Fatalf("len(IndividualVotes) = %d, want 3", len(page.IndividualVotes))
	}
	first := page.IndividualVotes[0]
	if first.GroupeUID != "GRP1" || first.Groupe != "G1" || first.ActeurUID != "PA1" || first.Depute != "Alice Martin" || first.Position != "pour" || !first.ParDelegation || first.NumPlace != "12" {
		t.Fatalf("first individual vote = %+v, want GRP1/PA1/pour delegated", first)
	}
	if page.IndividualVotes[1].ActeurUID != "PA2" || page.IndividualVotes[2].GroupeUID != "GRP2" {
		t.Fatalf("individual vote order = %+v, want grouped by preseance then position", page.IndividualVotes)
	}
}

func TestScrutinDetailPageReturnsLinkedReference(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertScrutin(t, s.db, testScrutin{
		UID:                     "linked",
		Numero:                  42,
		Date:                    "2024-02-02",
		Titre:                   "Vote sur un amendement",
		OrganeUID:               "ORG1",
		LinkedTextNum:           "2630",
		LinkedTextKind:          "projet-loi",
		LinkedTextURL:           "https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi",
		LinkedTextPDFURL:        "https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi.pdf",
		LinkedDossierRef:        "DLR5L17N1",
		LinkedDossierLibelle:    "Projet fixture",
		LinkedAmendementNum:     "301",
		LinkedAmendementTextNum: "2695",
		LinkedAmendementOrgane:  "AN",
		LinkedAmendementURL:     "https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/301.pdf",
		LinkedReferenceSource:   "titre",
	})

	page, err := s.ScrutinDetailPage(context.Background(), "linked")
	if err != nil {
		t.Fatalf("ScrutinDetailPage() error = %v", err)
	}

	if page.Scrutin.LinkedTextNum != "2630" || page.Scrutin.LinkedTextKind != "projet-loi" || page.Scrutin.LinkedTextURL != "https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi" || page.Scrutin.LinkedTextPDFURL != "https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi.pdf" || page.Scrutin.LinkedDossierRef != "DLR5L17N1" || page.Scrutin.LinkedDossierLibelle != "Projet fixture" || page.Scrutin.LinkedAmendementNum != "301" || page.Scrutin.LinkedAmendementTextNum != "2695" || page.Scrutin.LinkedAmendementOrgane != "AN" || page.Scrutin.LinkedAmendementURL != "https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/301.pdf" || page.Scrutin.LinkedReferenceSource != "titre" {
		t.Fatalf("linked reference = %+v, want text/dossier/amendment references", page.Scrutin)
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

	return New(testdb.Open(t, "test.sqlite"))
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

type testScrutin struct {
	UID                     string
	Numero                  int
	Legislature             int
	Date                    string
	Titre                   string
	OrganeUID               string
	Result                  string
	ResultLabel             string
	VoteType                string
	VoteTypeLabel           string
	Pour                    int
	Contre                  int
	LinkedTextNum           string
	LinkedTextKind          string
	LinkedTextURL           string
	LinkedTextPDFURL        string
	LinkedDossierRef        string
	LinkedDossierLibelle    string
	LinkedAmendementNum     string
	LinkedAmendementTextNum string
	LinkedAmendementOrgane  string
	LinkedAmendementURL     string
	LinkedReferenceSource   string
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
	legislature := scrutin.Legislature
	if legislature == 0 {
		legislature = 17
	}
	result := firstNonEmptyTest(scrutin.Result, "adopte")
	resultLabel := firstNonEmptyTest(scrutin.ResultLabel, "Adopte")
	voteType := firstNonEmptyTest(scrutin.VoteType, "SPO")
	voteTypeLabel := firstNonEmptyTest(scrutin.VoteTypeLabel, "Scrutin public ordinaire")
	pour := scrutin.Pour
	if pour == 0 {
		pour = 6
	}
	contre := scrutin.Contre
	if contre == 0 {
		contre = 3
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
) VALUES (?, ?, ?, ?, ?, ?, ?, 'simple', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'Gouvernement', 'Objet', 'Decompte', 10, 9, 5, 1, ?, ?, 0, 0, 'fixture.json')
`, scrutin.UID, scrutin.Numero, legislature, scrutin.OrganeUID, scrutin.Date, voteType, voteTypeLabel, result, resultLabel, scrutin.Titre, scrutin.LinkedTextNum, scrutin.LinkedTextKind, scrutin.LinkedTextURL, scrutin.LinkedTextPDFURL, scrutin.LinkedDossierRef, scrutin.LinkedDossierLibelle, scrutin.LinkedAmendementNum, scrutin.LinkedAmendementTextNum, scrutin.LinkedAmendementOrgane, scrutin.LinkedAmendementURL, scrutin.LinkedReferenceSource, pour, contre)
	if err != nil {
		t.Fatalf("insert scrutin %s: %v", scrutin.UID, err)
	}

	_, err = db.Exec(`INSERT INTO scrutin_search (uid, document) VALUES (?, ?)`, scrutin.UID, scrutin.Titre)
	if err != nil {
		t.Fatalf("insert scrutin search %s: %v", scrutin.UID, err)
	}
}

func hasFilterOption(options []ScrutinFilterOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

func firstNonEmptyTest(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func updateScrutinVoteCounts(t *testing.T, db *sql.DB, uid string, pour int, contre int) {
	t.Helper()
	_, err := db.Exec(`UPDATE scrutins SET pour = ?, contre = ? WHERE uid = ?`, pour, contre, uid)
	if err != nil {
		t.Fatalf("update scrutin counts %s: %v", uid, err)
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

func insertActeur(t *testing.T, db *sql.DB, uid, prenom, nom, alpha string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO acteurs (uid, prenom, nom, alpha) VALUES (?, ?, ?, ?)`, uid, prenom, nom, alpha)
	if err != nil {
		t.Fatalf("insert acteur %s: %v", uid, err)
	}
}

func insertIndividualVote(t *testing.T, db *sql.DB, scrutinUID, acteurUID, groupeUID, position string, parDelegation bool, numPlace string) {
	t.Helper()
	delegation := 0
	if parDelegation {
		delegation = 1
	}
	_, err := db.Exec(`
INSERT INTO votes (scrutin_uid, acteur_uid, groupe_uid, position, par_delegation, num_place)
VALUES (?, ?, ?, ?, ?, ?)
`, scrutinUID, acteurUID, groupeUID, position, delegation, numPlace)
	if err != nil {
		t.Fatalf("insert individual vote %s/%s: %v", scrutinUID, acteurUID, err)
	}
}

func uids(items []ScrutinListItem) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.UID)
	}
	return values
}
