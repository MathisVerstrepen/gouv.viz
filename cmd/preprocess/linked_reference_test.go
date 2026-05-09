package main

import "testing"

func TestExtractScrutinLinkedReference(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		objet         string
		wantText      string
		wantKind      string
		wantAmendment string
		wantSource    string
	}{
		{
			name:          "amendement and text in title",
			title:         "Scrutin sur l'amendement n° 301 au projet de loi (n° 2630)",
			wantText:      "2630",
			wantKind:      "projet-loi",
			wantAmendment: "301",
			wantSource:    "titre",
		},
		{
			name:          "amendment in object and text in title",
			title:         "Projet de loi (n° 1234)",
			objet:         "Vote sur l'amendement n° CL42 rectifié",
			wantText:      "1234",
			wantKind:      "projet-loi",
			wantAmendment: "CL42 rectifié",
			wantSource:    "titre+objet",
		},
		{
			name:       "global text vote",
			title:      "Vote solennel sur l'ensemble de la proposition de loi n° 987",
			wantText:   "987",
			wantKind:   "proposition-loi",
			wantSource: "titre",
		},
		{
			name:  "no reference",
			title: "Declaration du Gouvernement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractScrutinLinkedReference(tt.title, tt.objet)
			if got.TextNum != tt.wantText || got.TextKind != tt.wantKind || got.AmendementNum != tt.wantAmendment || got.Source != tt.wantSource {
				t.Fatalf("extractScrutinLinkedReference() = texte:%q kind:%q amendement:%q source:%q, want %q/%q/%q/%q", got.TextNum, got.TextKind, got.AmendementNum, got.Source, tt.wantText, tt.wantKind, tt.wantAmendment, tt.wantSource)
			}
		})
	}
}
