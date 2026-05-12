package components

import "testing"

func TestFormatDate(t *testing.T) {
	if got := formatDate("2024-07-18"); got != "18/07/2024" {
		t.Fatalf("formatDate(valid) = %q, want 18/07/2024", got)
	}
	if got := formatDate("not-a-date"); got != "not-a-date" {
		t.Fatalf("formatDate(invalid) = %q, want original value", got)
	}
}

func TestScrutinListURL(t *testing.T) {
	query := ScrutinsQuery{Search: " budget public ", Sort: "date_desc", Page: 3, PerPage: 25}

	if got := scrutinListURL(query, "date_desc", 1, "date_desc"); got != "/scrutins?q=+budget+public+" {
		t.Fatalf("scrutinListURL(default sort) = %q", got)
	}
	if got := scrutinListURL(query, "date_desc", 2, "closest"); got != "/scrutins?page=2&q=+budget+public+&sort=closest" {
		t.Fatalf("scrutinListURL(custom sort) = %q", got)
	}
	if got := scrutinListURL(query, "closest", 1, "date_desc"); got != "/scrutins?q=+budget+public+&sort=date_desc" {
		t.Fatalf("scrutinListURL(non-hardcoded default) = %q", got)
	}

	query = ScrutinsQuery{Legislature: 17, Result: "adopte", VoteType: "SPO", Organe: "PO800538", DateFrom: "2024-01-01", DateTo: "2024-12-31", CloseVotes: true}
	if got := scrutinListURL(query, "date_desc", 1, "date_desc"); got != "/scrutins?close_votes=1&date_from=2024-01-01&date_to=2024-12-31&legislature=17&organe=PO800538&result=adopte&vote_type=SPO" {
		t.Fatalf("scrutinListURL(filters) = %q", got)
	}
}

func TestPaginationPages(t *testing.T) {
	tests := []struct {
		name    string
		current int
		total   int
		want    []int
	}{
		{name: "single", current: 1, total: 1, want: []int{1}},
		{name: "small", current: 3, total: 5, want: []int{1, 2, 3, 4, 5}},
		{name: "middle", current: 5, total: 10, want: []int{1, 0, 3, 4, 5, 6, 7, 0, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := paginationPages(tt.current, tt.total)
			if len(got) != len(tt.want) {
				t.Fatalf("paginationPages() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("paginationPages() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestResultChipClass(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "adopte unaccented", code: "adopte", want: "result-chip result-chip--adopte"},
		{name: "adopte accented", code: " adopt\u00e9 ", want: "result-chip result-chip--adopte"},
		{name: "rejete unaccented", code: "rejete", want: "result-chip result-chip--rejete"},
		{name: "rejete accented", code: "rejet\u00e9", want: "result-chip result-chip--rejete"},
		{name: "unknown", code: "inconnu", want: "result-chip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resultChipClass(tt.code); got != tt.want {
				t.Fatalf("resultChipClass(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestCapitalizeFirstLetter(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: ""},
		{name: "lowercase", value: "projet de loi", want: "Projet de loi"},
		{name: "accented", value: "évaluation", want: "Évaluation"},
		{name: "already uppercase", value: "Budget", want: "Budget"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := capitalizeFirstLetter(tt.value); got != tt.want {
				t.Fatalf("capitalizeFirstLetter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetailHelpers(t *testing.T) {
	if got := ScrutinDetailURL("VT/1 2"); got != "/scrutins/VT%2F1%202" {
		t.Fatalf("ScrutinDetailURL() = %q", got)
	}
	if got := DeputyDetailURL("PA/1 2"); got != "/deputes/PA%2F1%202" {
		t.Fatalf("DeputyDetailURL() = %q", got)
	}
	if got := fallbackText(""); got != "Non renseigné" {
		t.Fatalf("fallbackText(empty) = %q", got)
	}
	if got := fallbackText("Valeur"); got != "Valeur" {
		t.Fatalf("fallbackText(value) = %q", got)
	}
	if got := numberOrEmpty(0); got != "" {
		t.Fatalf("numberOrEmpty(0) = %q", got)
	}
	if got := numberOrEmpty(12); got != "12" {
		t.Fatalf("numberOrEmpty(12) = %q", got)
	}
}

func TestDeputyDetailHelpers(t *testing.T) {
	page := DeputyDetailPage{
		Stats: []DeputyVoteStat{
			{Legislature: 17, TotalVotes: 3, Pour: 2, Contre: 1},
			{Legislature: 16, TotalVotes: 2, Pour: 1, Contre: 0},
		},
	}
	if got := deputyTotalVotes(page); got != 5 {
		t.Fatalf("deputyTotalVotes() = %d, want 5", got)
	}
	if got := deputyTotalPour(page.Stats); got != 3 {
		t.Fatalf("deputyTotalPour() = %d, want 3", got)
	}
	if got := deputyTotalContre(page.Stats); got != 1 {
		t.Fatalf("deputyTotalContre() = %d, want 1", got)
	}
	if got := latestMandatLabel([]DeputyMandat{{LibQualite: "Députée", TypeOrgane: "ASSEMBLEE"}}); got != "Députée" {
		t.Fatalf("latestMandatLabel() = %q, want Députée", got)
	}
	if got := deputyBirthPlace(DeputyDetailData{VilleNaissance: "Lille", DepNaissance: "59", PaysNaissance: "France"}); got != "Lille · 59 · France" {
		t.Fatalf("deputyBirthPlace() = %q", got)
	}
	photoPage := DeputyDetailPage{
		Deputy:  DeputyDetailData{UID: "PA794598"},
		Mandats: []DeputyMandat{{Legislature: 17}},
	}
	if got := deputyPhotoURL(photoPage); got != "https://www.assemblee-nationale.fr/dyn/static/tribun/17/photos/carre/794598.jpg" {
		t.Fatalf("deputyPhotoURL() = %q", got)
	}
	if got := deputyPhotoID("PA/1"); got != "" {
		t.Fatalf("deputyPhotoID() = %q, want empty for non-numeric UID", got)
	}
	if got := deputyRemainingMandatsLabel(1); got != "Voir 1 autre mandat" {
		t.Fatalf("deputyRemainingMandatsLabel(1) = %q", got)
	}
	if got := deputyRemainingMandatsLabel(3); got != "Voir 3 autres mandats" {
		t.Fatalf("deputyRemainingMandatsLabel(3) = %q", got)
	}
	votesPage := DeputyDetailPage{Deputy: DeputyDetailData{UID: "PA/1"}, Query: DeputyDetailQuery{VotesPage: 2, VotesSort: "date_desc"}, VotesTotalResults: 75, VotesStartItem: 51, VotesEndItem: 75}
	if got := deputyVotesPageURL(votesPage, 2); got != "/deputes/PA%2F1?votes_page=2#depute-votes" {
		t.Fatalf("deputyVotesPageURL() = %q", got)
	}
	if got := deputyVotesPageURL(votesPage, 1); got != "/deputes/PA%2F1#depute-votes" {
		t.Fatalf("deputyVotesPageURL(first) = %q", got)
	}
	filteredVotesPage := DeputyDetailPage{Deputy: DeputyDetailData{UID: "PA/1"}, Query: DeputyDetailQuery{VotesSearch: "budget", VotesPosition: "pour", VotesSort: "date_asc"}}
	if got := deputyVotesPageURL(filteredVotesPage, 3); got != "/deputes/PA%2F1?votes_page=3&votes_position=pour&votes_q=budget&votes_sort=date_asc#depute-votes" {
		t.Fatalf("deputyVotesPageURL(filtered) = %q", got)
	}
	if got := deputyVotesSummary(votesPage); got != "51-75 sur 75 votes nominatifs." {
		t.Fatalf("deputyVotesSummary() = %q", got)
	}
	if !hasDeputyVoteControls(filteredVotesPage.Query) {
		t.Fatal("hasDeputyVoteControls(filtered) = false, want true")
	}
	if got := deputyVoteResultLabel(DeputyVote{SortLibelle: "l'Assemblée nationale a adopté"}); got != "Adopté" {
		t.Fatalf("deputyVoteResultLabel(adopted) = %q", got)
	}
	if got := deputyVoteResultLabel(DeputyVote{SortLibelle: "L’Assemblée nationale n’a pas adopté"}); got != "Non adopté" {
		t.Fatalf("deputyVoteResultLabel(not adopted) = %q", got)
	}
	group := DeputyMandatOrgane{UID: "PO800538", CodeType: "GP", Libelle: "Ensemble pour la République", LibelleAbrev: "EPR"}
	if got := groupLogoURL(group); got != "/assets/img/groups/EPR.png" {
		t.Fatalf("groupLogoURL() = %q", got)
	}
	if got := groupLogoURL(DeputyMandatOrgane{CodeType: "GP", LibelleAbrev: "inconnu"}); got != "" {
		t.Fatalf("groupLogoURL(unknown) = %q, want empty", got)
	}
	if got := deputyCurrentGroup([]DeputyMandat{{Organes: []DeputyMandatOrgane{{CodeType: "ORG", Libelle: "Commission"}, group}}}); got.UID != group.UID {
		t.Fatalf("deputyCurrentGroup() = %+v, want %+v", got, group)
	}
}

func TestIndividualVoteGroups(t *testing.T) {
	votes := []ScrutinIndividualVote{
		{GroupeUID: "G1", Groupe: "Groupe un", ActeurUID: "PA1", Position: "pour"},
		{GroupeUID: "G1", Groupe: "Groupe un", ActeurUID: "PA2", Position: "contre"},
		{GroupeUID: "G2", Groupe: "Groupe deux", ActeurUID: "PA3", Position: "abstention"},
	}

	groups := individualVoteGroups(votes)
	if len(groups) != 2 {
		t.Fatalf("len(individualVoteGroups) = %d, want 2", len(groups))
	}
	if groups[0].Groupe != "Groupe un" || len(groups[0].Votes) != 2 || groups[1].Groupe != "Groupe deux" || len(groups[1].Votes) != 1 {
		t.Fatalf("individualVoteGroups() = %+v", groups)
	}
}

func TestIndividualVotePositionHelpers(t *testing.T) {
	if got := individualVotePositionLabel("non_votant_volontaire"); got != "Non votant volontaire" {
		t.Fatalf("individualVotePositionLabel() = %q", got)
	}
	if got := individualVotePositionClass("contre"); got != "vote-position vote-position--contre" {
		t.Fatalf("individualVotePositionClass() = %q", got)
	}
	if got := delegationLabel(true); got != "Oui" {
		t.Fatalf("delegationLabel(true) = %q", got)
	}
}

func TestOfficialLinkedTextURL(t *testing.T) {
	if got := officialLinkedTextURL(ScrutinDetailData{LinkedTextURL: "https://example.test/text"}); got != "https://example.test/text" {
		t.Fatalf("officialLinkedTextURL(stored) = %q", got)
	}

	scrutin := ScrutinDetailData{Legislature: 17, LinkedTextNum: "2630", LinkedTextKind: "projet-loi"}
	if got := officialLinkedTextURL(scrutin); got != "https://www.assemblee-nationale.fr/dyn/17/textes/l17b2630_projet-loi" {
		t.Fatalf("officialLinkedTextURL() = %q", got)
	}

	shortNum := ScrutinDetailData{Legislature: 17, LinkedTextNum: "324", LinkedTextKind: "projet-loi"}
	if got := officialLinkedTextURL(shortNum); got != "https://www.assemblee-nationale.fr/dyn/17/textes/l17b0324_projet-loi" {
		t.Fatalf("officialLinkedTextURL(short num) = %q", got)
	}

	if got := officialLinkedTextURL(ScrutinDetailData{Legislature: 17, LinkedTextNum: "2630"}); got != "" {
		t.Fatalf("officialLinkedTextURL(missing kind) = %q, want empty", got)
	}
}

func TestOfficialLinkedTextPDFURL(t *testing.T) {
	scrutin := ScrutinDetailData{LinkedTextPDFURL: "https://www.assemblee-nationale.fr/dyn/17/textes/l17t0278_texte-adopte-seance.pdf"}
	if got := officialLinkedTextPDFURL(scrutin); got != scrutin.LinkedTextPDFURL {
		t.Fatalf("officialLinkedTextPDFURL() = %q", got)
	}
}

func TestOfficialDossierURL(t *testing.T) {
	scrutin := ScrutinDetailData{Legislature: 17, LinkedDossierRef: "DLR5L17N54083"}
	if got := officialDossierURL(scrutin); got != "https://www.assemblee-nationale.fr/dyn/17/dossiers/DLR5L17N54083" {
		t.Fatalf("officialDossierURL() = %q", got)
	}

	if got := officialDossierURL(ScrutinDetailData{Legislature: 17}); got != "" {
		t.Fatalf("officialDossierURL(missing ref) = %q, want empty", got)
	}
}

func TestOfficialAmendementURL(t *testing.T) {
	scrutin := ScrutinDetailData{LinkedAmendementURL: "https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/365.pdf"}
	if got := officialAmendementURL(scrutin); got != "https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/365.pdf" {
		t.Fatalf("officialAmendementURL() = %q", got)
	}

	if got := officialAmendementURL(ScrutinDetailData{}); got != "" {
		t.Fatalf("officialAmendementURL(empty) = %q, want empty", got)
	}
}
