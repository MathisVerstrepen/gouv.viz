package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryDossierResolverResolvesCMPTextPDF(t *testing.T) {
	root := filepath.Join(t.TempDir(), "dossiers")
	writeTestDossiersDirectory(t, root)

	resolver, err := newDirectoryDossierResolver(root)
	if err != nil {
		t.Fatalf("newDirectoryDossierResolver() error = %v", err)
	}

	got := resolver.Resolve("DLR5L17N52985", "l'ensemble du projet de loi relatif à la lutte contre les fraudes sociales et fiscales (texte de la commission mixte paritaire).", "", "adopté")
	if got.Text.Num != "2701" || got.Text.Kind != "texte-adopte-commission" || got.Text.PDFURL != "https://www.assemblee-nationale.fr/dyn/17/textes/l17b2701_texte-adopte-commission.pdf" {
		t.Fatalf("Resolve(CMP) = %+v, want text 2701 commission PDF", got)
	}
}

func TestDirectoryDossierResolverMatchesMissingScrutinDossierByTitle(t *testing.T) {
	root := filepath.Join(t.TempDir(), "dossiers")
	writeTestDossiersDirectory(t, root)

	resolver, err := newDirectoryDossierResolver(root)
	if err != nil {
		t.Fatalf("newDirectoryDossierResolver() error = %v", err)
	}

	got := resolver.Resolve("", "la première partie du projet de loi de finances pour 2025 (première lecture).", "", "rejeté")
	if got.DossierRef != "DLR5L17N50198" || got.Text.Num != "324" || got.Text.Kind != "projet-loi" {
		t.Fatalf("Resolve(missing dossier) = %+v, want PLF 2025 original text", got)
	}
}

func TestDossiersDirectoryPathUsesDefaultWhenPresent(t *testing.T) {
	rawDir := t.TempDir()
	root := filepath.Join(rawDir, "dossiers")
	writeTestDossiersDirectory(t, root)

	got, ok := dossiersDirectoryPath(rawDir)
	if !ok || got != root {
		t.Fatalf("dossiersDirectoryPath() = %q/%v, want default directory", got, ok)
	}
}

func TestSortOfficialTextReferencesIsDeterministic(t *testing.T) {
	refs := []officialTextReference{
		{Num: "279", Kind: "texte-adopte-seance", URL: "seance-279"},
		{Num: "2701", Kind: "texte-adopte-commission", URL: "commission-2701"},
		{Num: "2115", Kind: "projet-loi", URL: "projet-2115"},
		{Num: "2250", Kind: "texte-adopte-commission", URL: "commission-2250"},
	}

	sortOfficialTextReferences(refs)

	want := []string{"projet-2115", "commission-2250", "commission-2701", "seance-279"}
	for i := range want {
		if refs[i].URL != want[i] {
			t.Fatalf("sorted refs[%d] = %q, want %q (all refs: %+v)", i, refs[i].URL, want[i], refs)
		}
	}
}

func writeTestDossiersDirectory(t *testing.T, root string) {
	t.Helper()

	files := map[string]string{
		"json/dossierParlementaire/DLR5L17N52985.json": `{
			"dossierParlementaire": {
				"uid": "DLR5L17N52985",
				"titreDossier": {"titre": "Projet de loi relatif à la lutte contre les fraudes sociales et fiscales"},
				"actesLegislatifs": {"acteLegislatif": [
					{"texteAssocie": "PRJLANR5L17B2115"},
					{"texteAssocie": "PRJLANR5L17BTC2250"},
					{"texteAssocie": "PRJLANR5L17BTC2701"},
					{"texteAssocie": "PRJLANR5L17BTA0279"}
				]}
			}
		}`,
		"json/dossierParlementaire/DLR5L17N50198.json": `{
			"dossierParlementaire": {
				"uid": "DLR5L17N50198",
				"titreDossier": {"titre": "Projet de loi de finances pour 2025"},
				"actesLegislatifs": {"acteLegislatif": [
					{"texteAssocie": "PRJLANR5L17B0324"},
					{"texteAssocie": "PRJLANR5L17BTC0468"}
				]}
			}
		}`,
	}

	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create directory for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}
