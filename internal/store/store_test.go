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

func TestDeputyDetailPageReturnsProfileMandatsStatsAndVotes(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertOrgane(t, s.db, "GRP1", "Groupe un", "G1", 2)
	insertActeur(t, s.db, "PA1", "Alice", "Martin", "MARTIN")
	if _, err := s.db.Exec(`
UPDATE acteurs
SET civilite = 'Mme', date_naissance = '1980-01-02', ville_naissance = 'Lille', dep_naissance = '59', pays_naissance = 'France', profession = 'Juriste', uri_hatvp = 'https://hatvp.example/pa1', source_file = 'acteur.json'
WHERE uid = 'PA1'
`); err != nil {
		t.Fatalf("update acteur details: %v", err)
	}
	insertMandat(t, s.db, "M1", "PA1", 17, "ASSEMBLEE", "2024-01-01", "", "Députée", true)
	insertMandatOrgane(t, s.db, "M1", "GRP1")
	insertScrutin(t, s.db, testScrutin{UID: "old", Numero: 1, Date: "2024-01-10", Titre: "Ancien vote", OrganeUID: "ORG1", Result: "rejete", ResultLabel: "Rejeté"})
	insertScrutin(t, s.db, testScrutin{UID: "new", Numero: 2, Date: "2024-02-10", Titre: "Nouveau vote", OrganeUID: "ORG1", Result: "adopte", ResultLabel: "Adopté"})
	insertIndividualVoteWithMandat(t, s.db, "old", "PA1", "M1", "GRP1", "contre", false, "")
	insertIndividualVoteWithMandat(t, s.db, "new", "PA1", "M1", "GRP1", "pour", true, "12")
	if _, err := s.db.Exec(`INSERT INTO acteur_vote_stats (acteur_uid, legislature, total_votes, pour, contre, abstentions, non_votants) VALUES ('PA1', 17, 2, 1, 1, 0, 0)`); err != nil {
		t.Fatalf("insert actor vote stats: %v", err)
	}

	page, err := s.DeputyDetailPage(context.Background(), "PA1", DeputyDetailQuery{})
	if err != nil {
		t.Fatalf("DeputyDetailPage() error = %v", err)
	}

	if page.Deputy.DisplayName != "Alice Martin" || page.Deputy.Civilite != "Mme" || page.Deputy.URIHATVP != "https://hatvp.example/pa1" {
		t.Fatalf("deputy = %+v, want detailed profile", page.Deputy)
	}
	if len(page.Mandats) != 1 || page.Mandats[0].UID != "M1" || !page.Mandats[0].NominPrincipale || len(page.Mandats[0].Organes) != 1 || page.Mandats[0].Organes[0].UID != "GRP1" {
		t.Fatalf("mandats = %+v, want M1 with GRP1", page.Mandats)
	}
	if len(page.Stats) != 1 || page.Stats[0].TotalVotes != 2 || page.Stats[0].Pour != 1 || page.Stats[0].Contre != 1 {
		t.Fatalf("stats = %+v, want 2 total votes", page.Stats)
	}
	if len(page.Votes) != 2 || page.Votes[0].ScrutinUID != "new" || page.Votes[0].Position != "pour" || !page.Votes[0].ParDelegation || page.Votes[0].Groupe != "G1" || page.Votes[1].ScrutinUID != "old" {
		t.Fatalf("votes = %+v, want reverse chronological votes with group", page.Votes)
	}
	if page.Query.VotesPage != 1 || page.VotesTotalResults != 2 || page.VotesTotalPages != 1 || page.VotesStartItem != 1 || page.VotesEndItem != 2 {
		t.Fatalf("vote pagination = query:%+v total:%d pages:%d start:%d end:%d", page.Query, page.VotesTotalResults, page.VotesTotalPages, page.VotesStartItem, page.VotesEndItem)
	}
}

func TestDeputyDetailPagePaginatesVotes(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertOrgane(t, s.db, "GRP1", "Groupe un", "G1", 2)
	insertActeur(t, s.db, "PA1", "Alice", "Martin", "MARTIN")
	insertScrutin(t, s.db, testScrutin{UID: "vote-1", Numero: 1, Date: "2024-01-01", Titre: "Premier vote", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "vote-2", Numero: 2, Date: "2024-01-02", Titre: "Deuxième vote", OrganeUID: "ORG1"})
	insertScrutin(t, s.db, testScrutin{UID: "vote-3", Numero: 3, Date: "2024-01-03", Titre: "Troisième vote", OrganeUID: "ORG1"})
	insertIndividualVote(t, s.db, "vote-1", "PA1", "GRP1", "pour", false, "")
	insertIndividualVote(t, s.db, "vote-2", "PA1", "GRP1", "contre", false, "")
	insertIndividualVote(t, s.db, "vote-3", "PA1", "GRP1", "abstention", false, "")

	page, err := s.DeputyDetailPage(context.Background(), "PA1", DeputyDetailQuery{VotesPage: 2, VotesPerPage: 1})
	if err != nil {
		t.Fatalf("DeputyDetailPage() error = %v", err)
	}

	if len(page.Votes) != 1 || page.Votes[0].ScrutinUID != "vote-2" {
		t.Fatalf("votes page = %+v, want second newest vote", page.Votes)
	}
	if page.Query.VotesPage != 2 || page.Query.VotesPerPage != 1 || page.VotesTotalResults != 3 || page.VotesTotalPages != 3 || page.VotesStartItem != 2 || page.VotesEndItem != 2 {
		t.Fatalf("vote pagination = query:%+v total:%d pages:%d start:%d end:%d", page.Query, page.VotesTotalResults, page.VotesTotalPages, page.VotesStartItem, page.VotesEndItem)
	}

	page, err = s.DeputyDetailPage(context.Background(), "PA1", DeputyDetailQuery{VotesSearch: "deuxième", VotesPosition: "contre", VotesSort: "date_asc", VotesPerPage: 10})
	if err != nil {
		t.Fatalf("DeputyDetailPage(filtered) error = %v", err)
	}
	if len(page.Votes) != 1 || page.Votes[0].ScrutinUID != "vote-2" || page.VotesTotalResults != 1 {
		t.Fatalf("filtered votes = %+v total=%d, want vote-2 only", page.Votes, page.VotesTotalResults)
	}
	if page.Query.VotesSearch != "deuxième" || page.Query.VotesPosition != "contre" || page.Query.VotesSort != "date_asc" {
		t.Fatalf("normalized filter query = %+v", page.Query)
	}

	page, err = s.DeputyDetailPage(context.Background(), "PA1", DeputyDetailQuery{VotesSort: "date_asc", VotesPerPage: 10})
	if err != nil {
		t.Fatalf("DeputyDetailPage(sorted) error = %v", err)
	}
	if len(page.Votes) != 3 || page.Votes[0].ScrutinUID != "vote-1" || page.Votes[2].ScrutinUID != "vote-3" {
		t.Fatalf("date ascending votes = %+v", page.Votes)
	}
}

func TestDeputyDetailPageReturnsErrNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.DeputyDetailPage(context.Background(), "missing", DeputyDetailQuery{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("DeputyDetailPage() error = %v, want ErrNotFound", err)
	}
}

func TestDeputiesPageAppliesSearchFiltersSortAndPagination(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "GRP1", "Groupe un", "G1", 1)
	insertOrgane(t, s.db, "GRP2", "Groupe deux", "G2", 2)
	insertActeur(t, s.db, "PA1", "Alice", "Martin", "MARTIN")
	insertActeur(t, s.db, "PA2", "Bruno", "Bernard", "BERNARD")
	insertActeur(t, s.db, "PA3", "Claire", "Durand", "DURAND")
	if _, err := s.db.Exec(`UPDATE acteurs SET profession = 'Juriste' WHERE uid = 'PA1'`); err != nil {
		t.Fatalf("update actor profession: %v", err)
	}
	insertMandat(t, s.db, "M1", "PA1", 17, "ASSEMBLEE", "2024-01-01", "", "Députée", true)
	insertMandatOrgane(t, s.db, "M1", "GRP1")
	insertMandat(t, s.db, "M2", "PA2", 17, "ASSEMBLEE", "2024-01-01", "", "Député", true)
	insertMandatOrgane(t, s.db, "M2", "GRP2")
	insertMandat(t, s.db, "M3", "PA3", 16, "ASSEMBLEE", "2023-01-01", "", "Députée", true)
	insertMandatOrgane(t, s.db, "M3", "GRP1")
	insertActorVoteStats(t, s.db, "PA1", 17, 12, 7, 3, 1, 1)
	insertActorVoteStats(t, s.db, "PA2", 17, 3, 1, 1, 1, 0)
	insertActorVoteStats(t, s.db, "PA3", 16, 8, 2, 4, 1, 1)

	page, err := s.DeputiesPage(context.Background(), DeputiesQuery{Search: "juriste", Legislature: 17, Group: "GRP1", Page: 1, PerPage: 25})
	if err != nil {
		t.Fatalf("DeputiesPage() error = %v", err)
	}
	if got := deputyUIDs(page.Deputies); len(got) != 1 || got[0] != "PA1" {
		t.Fatalf("deputies = %v, want [PA1]", got)
	}
	if page.Deputies[0].Group != "G1" || page.Deputies[0].TotalVotes != 12 || page.Deputies[0].Pour != 7 {
		t.Fatalf("deputy summary = %+v, want group and vote stats", page.Deputies[0])
	}
	if !hasDeputyFilterOption(page.FilterOptions.Groups, "GRP1") || len(page.FilterOptions.Legislatures) == 0 || page.FilterOptions.Legislatures[0].Value != "17" {
		t.Fatalf("filter options = %+v, want groups and legislatures", page.FilterOptions)
	}

	page, err = s.DeputiesPage(context.Background(), DeputiesQuery{Sort: "votes_desc", Page: 1, PerPage: 2})
	if err != nil {
		t.Fatalf("DeputiesPage(sorted) error = %v", err)
	}
	if got := deputyUIDs(page.Deputies); len(got) != 2 || got[0] != "PA1" || got[1] != "PA3" {
		t.Fatalf("sorted deputies = %v, want [PA1 PA3]", got)
	}
	if page.TotalResults != 3 || page.TotalPages != 2 || page.StartItem != 1 || page.EndItem != 2 {
		t.Fatalf("pagination = total:%d pages:%d start:%d end:%d", page.TotalResults, page.TotalPages, page.StartItem, page.EndItem)
	}

	page, err = s.DeputiesPage(context.Background(), DeputiesQuery{Page: 99, PerPage: 2})
	if err != nil {
		t.Fatalf("DeputiesPage(clamped) error = %v", err)
	}
	if page.Query.Page != 2 || page.StartItem != 3 || page.EndItem != 3 {
		t.Fatalf("clamped pagination = query:%+v start:%d end:%d", page.Query, page.StartItem, page.EndItem)
	}
}

func TestPoliticalGroupDetailPageReturnsStatsDeputiesAndVotes(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)
	insertOrgane(t, s.db, "GRP1", "Groupe un", "G1", 2)
	insertActeur(t, s.db, "PA1", "Alice", "Martin", "MARTIN")
	insertActeur(t, s.db, "PA2", "Bruno", "Bernard", "BERNARD")
	insertMandat(t, s.db, "M1", "PA1", 17, "ASSEMBLEE", "2024-01-01", "", "Députée", true)
	insertMandatOrgane(t, s.db, "M1", "GRP1")
	insertMandat(t, s.db, "M2", "PA2", 17, "ASSEMBLEE", "2024-01-02", "", "Député", true)
	insertMandatOrgane(t, s.db, "M2", "GRP1")
	insertScrutin(t, s.db, testScrutin{UID: "old", Numero: 1, Date: "2024-01-10", Titre: "Ancien vote", OrganeUID: "ORG1", Result: "rejete", ResultLabel: "Rejeté"})
	insertScrutin(t, s.db, testScrutin{UID: "new", Numero: 2, Date: "2024-02-10", Titre: "Nouveau vote", OrganeUID: "ORG1", Result: "adopte", ResultLabel: "Adopté"})
	insertGroupVote(t, s.db, "old", "GRP1", "contre")
	insertGroupVote(t, s.db, "new", "GRP1", "pour")
	insertGroupVoteStats(t, s.db, "GRP1", 17, 2, 1, 1, 0, 0)

	page, err := s.PoliticalGroupDetailPage(context.Background(), "GRP1", PoliticalGroupDetailQuery{VotesPage: 1, VotesPerPage: 1})
	if err != nil {
		t.Fatalf("PoliticalGroupDetailPage() error = %v", err)
	}

	if page.Group.UID != "GRP1" || page.Group.LibelleAbrege != "G1" || page.Group.CodeType != "GP" {
		t.Fatalf("group = %+v, want GRP1", page.Group)
	}
	if len(page.Stats) != 1 || page.Stats[0].TotalScrutins != 2 || page.Stats[0].Pour != 1 || page.Stats[0].Contre != 1 {
		t.Fatalf("stats = %+v, want aggregate group stats", page.Stats)
	}
	if got := politicalGroupDeputyUIDs(page.Deputies); len(got) != 2 || got[0] != "PA2" || got[1] != "PA1" {
		t.Fatalf("deputies = %v, want alphabetical [PA2 PA1]", got)
	}
	if len(page.Votes) != 1 || page.Votes[0].ScrutinUID != "new" || page.Votes[0].PositionMajoritaire != "pour" {
		t.Fatalf("votes = %+v, want newest group vote", page.Votes)
	}
	if page.VotesTotalResults != 2 || page.VotesTotalPages != 2 || page.VotesStartItem != 1 || page.VotesEndItem != 1 {
		t.Fatalf("vote pagination = total:%d pages:%d start:%d end:%d", page.VotesTotalResults, page.VotesTotalPages, page.VotesStartItem, page.VotesEndItem)
	}

	page, err = s.PoliticalGroupDetailPage(context.Background(), "GRP1", PoliticalGroupDetailQuery{VotesSearch: "ancien", VotesPosition: "contre", VotesSort: "date_asc", VotesPerPage: 10})
	if err != nil {
		t.Fatalf("PoliticalGroupDetailPage(filtered) error = %v", err)
	}
	if len(page.Votes) != 1 || page.Votes[0].ScrutinUID != "old" || page.VotesTotalResults != 1 {
		t.Fatalf("filtered votes = %+v total=%d, want old only", page.Votes, page.VotesTotalResults)
	}
	if page.Query.VotesSearch != "ancien" || page.Query.VotesPosition != "contre" || page.Query.VotesSort != "date_asc" {
		t.Fatalf("normalized vote query = %+v", page.Query)
	}
}

func TestPoliticalGroupDetailPageReturnsErrNotFound(t *testing.T) {
	s := newTestStore(t)
	insertOrgane(t, s.db, "ORG1", "Commission", "COM", 1)

	_, err := s.PoliticalGroupDetailPage(context.Background(), "ORG1", PoliticalGroupDetailQuery{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("PoliticalGroupDetailPage() error = %v, want ErrNotFound", err)
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
	_, err := db.Exec(`INSERT INTO organes (uid, code_type, libelle, libelle_abrege, libelle_abrev, preseance) VALUES (?, ?, ?, ?, ?, ?)`, uid, organeCodeType(uid), libelle, libelleAbrege, libelleAbrege, preseance)
	if err != nil {
		t.Fatalf("insert organe %s: %v", uid, err)
	}
}

func organeCodeType(uid string) string {
	if strings.HasPrefix(uid, "GRP") {
		return "GP"
	}
	return "ORG"
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

func hasDeputyFilterOption(options []DeputyFilterOption, value string) bool {
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

func insertActorVoteStats(t *testing.T, db *sql.DB, acteurUID string, legislature int, total, pour, contre, abstentions, nonVotants int) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO acteur_vote_stats (acteur_uid, legislature, total_votes, pour, contre, abstentions, non_votants) VALUES (?, ?, ?, ?, ?, ?, ?)`, acteurUID, legislature, total, pour, contre, abstentions, nonVotants)
	if err != nil {
		t.Fatalf("insert actor vote stats %s: %v", acteurUID, err)
	}
}

func insertGroupVoteStats(t *testing.T, db *sql.DB, groupeUID string, legislature int, total, pour, contre, abstentions, nonVotants int) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO groupe_vote_stats (groupe_uid, legislature, total_scrutins, pour, contre, abstentions, non_votants) VALUES (?, ?, ?, ?, ?, ?, ?)`, groupeUID, legislature, total, pour, contre, abstentions, nonVotants)
	if err != nil {
		t.Fatalf("insert group vote stats %s: %v", groupeUID, err)
	}
}

func insertIndividualVote(t *testing.T, db *sql.DB, scrutinUID, acteurUID, groupeUID, position string, parDelegation bool, numPlace string) {
	t.Helper()
	insertIndividualVoteWithMandat(t, db, scrutinUID, acteurUID, "", groupeUID, position, parDelegation, numPlace)
}

func insertIndividualVoteWithMandat(t *testing.T, db *sql.DB, scrutinUID, acteurUID, mandatUID, groupeUID, position string, parDelegation bool, numPlace string) {
	t.Helper()
	delegation := 0
	if parDelegation {
		delegation = 1
	}
	_, err := db.Exec(`
INSERT INTO votes (scrutin_uid, acteur_uid, mandat_uid, groupe_uid, position, par_delegation, num_place)
VALUES (?, ?, ?, ?, ?, ?, ?)
`, scrutinUID, acteurUID, nullStringTest(mandatUID), groupeUID, position, delegation, numPlace)
	if err != nil {
		t.Fatalf("insert individual vote %s/%s: %v", scrutinUID, acteurUID, err)
	}
}

func insertMandat(t *testing.T, db *sql.DB, uid, acteurUID string, legislature int, typeOrgane, dateDebut, dateFin, libQualite string, nominPrincipale bool) {
	t.Helper()
	nomin := 0
	if nominPrincipale {
		nomin = 1
	}
	_, err := db.Exec(`
INSERT INTO mandats (uid, acteur_uid, legislature, type_organe, date_debut, date_fin, nomin_principale, lib_qualite)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, uid, acteurUID, legislature, typeOrgane, dateDebut, nullStringTest(dateFin), nomin, libQualite)
	if err != nil {
		t.Fatalf("insert mandat %s: %v", uid, err)
	}
}

func insertMandatOrgane(t *testing.T, db *sql.DB, mandatUID, organeUID string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO mandat_organes (mandat_uid, organe_uid) VALUES (?, ?)`, mandatUID, organeUID)
	if err != nil {
		t.Fatalf("insert mandat organe %s/%s: %v", mandatUID, organeUID, err)
	}
}

func nullStringTest(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func uids(items []ScrutinListItem) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.UID)
	}
	return values
}

func deputyUIDs(items []DeputyListItem) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.UID)
	}
	return values
}

func politicalGroupDeputyUIDs(items []PoliticalGroupDeputy) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.UID)
	}
	return values
}
