package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryAmendementResolverResolvesPublicSessionPDF(t *testing.T) {
	root := filepath.Join(t.TempDir(), "amendements")
	writeTestAmendementsDirectory(t, root)

	resolver, err := newDirectoryAmendementResolver(root)
	if err != nil {
		t.Fatalf("newDirectoryAmendementResolver() error = %v", err)
	}

	got, err := resolver.Resolve(17, "DLR5L17N54083", "365", publicSessionOrganeRef, "RUANR5L17S2026IDS30540")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.TextNum != "2695" || got.Organe != "AN" || got.OrganeRef != publicSessionOrganeRef || got.AmendementNum != "365" || got.URL != "https://www.assemblee-nationale.fr/dyn/17/amendements/2695/AN/365.pdf" {
		t.Fatalf("Resolve() = %+v, want AN amendment PDF on text 2695", got)
	}
}

func TestChooseOfficialAmendementReferenceFallsBackToExactOrgane(t *testing.T) {
	candidates := []officialAmendementReference{
		{TextNum: "2630", Organe: "CION_DEF", OrganeRef: "PO59046", AmendementNum: "DN365", URL: "commission.pdf"},
		{TextNum: "2695", Organe: "AN", OrganeRef: publicSessionOrganeRef, AmendementNum: "365", URL: "public.pdf"},
	}

	got := chooseOfficialAmendementReference(candidates, "365", "PO59046", "")
	if got.TextNum != "2630" || got.Organe != "CION_DEF" || got.AmendementNum != "DN365" || got.URL != "commission.pdf" {
		t.Fatalf("choose exact organe = %+v, want CION_DEF/DN365 on text 2630", got)
	}
}

func TestAmendementDirectoryPathUsesDefaultWhenPresent(t *testing.T) {
	rawDir := t.TempDir()
	root := filepath.Join(rawDir, "amendements")
	writeTestAmendementsDirectory(t, root)

	got, ok := amendementsDirectoryPath(rawDir)
	if !ok || got != root {
		t.Fatalf("amendementsDirectoryPath() = %q/%v, want default directory", got, ok)
	}
}

func writeTestAmendementsDirectory(t *testing.T, root string) {
	t.Helper()

	files := map[string]string{
		"json/DLR5L17N54083/PRJLANR5L17BTC2695/AMANR5L17PO838901BTC2695P0D1N000365.json": `{
			"amendement": {
				"uid": "AMANR5L17PO838901BTC2695P0D1N000365",
				"legislature": "17",
				"identification": {"numeroLong": "365", "prefixeOrganeExamen": "AN"},
				"texteLegislatifRef": "PRJLANR5L17BTC2695",
				"seanceDiscussionRef": "RUANR5L17S2026IDS30540"
			}
		}`,
		"json/DLR5L17N54083/PRJLANR5L17B2630/AMANR5L17PO59046B2630P0D1N000365.json": `{
			"amendement": {
				"uid": "AMANR5L17PO59046B2630P0D1N000365",
				"legislature": "17",
				"identification": {"numeroLong": "DN365", "prefixeOrganeExamen": "CION_DEF"},
				"texteLegislatifRef": "PRJLANR5L17B2630"
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
