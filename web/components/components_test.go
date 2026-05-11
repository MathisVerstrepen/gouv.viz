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

func TestDetailHelpers(t *testing.T) {
	if got := ScrutinDetailURL("VT/1 2"); got != "/scrutins/VT%2F1%202" {
		t.Fatalf("ScrutinDetailURL() = %q", got)
	}
	if got := fallbackText(""); got != "Non renseigne" {
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
