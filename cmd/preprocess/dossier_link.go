package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type officialTextReference struct {
	Num    string
	Kind   string
	URL    string
	PDFURL string
}

type officialDossierReference struct {
	DossierRef     string
	DossierLibelle string
	Text           officialTextReference
	Texts          []officialTextReference
}

type dossierLinkResolver interface {
	Resolve(dossierRef, title, object, sortCode string) officialDossierReference
}

type directoryDossierResolver struct {
	dossiers map[string]officialDossierReference
	byTitle  []officialDossierReference
}

var (
	texteOriginalRE   = regexp.MustCompile(`^(PION|PRJL)ANR5L(\d+)B0*(\d+)$`)
	texteCommissionRE = regexp.MustCompile(`^(PION|PRJL)ANR5L(\d+)BTC0*(\d+)$`)
	texteAdopteRE     = regexp.MustCompile(`^(PION|PRJL)ANR5L(\d+)(?:BTA|TAP)0*(\d+)$`)
)

func newDirectoryDossierResolver(root string) (*directoryDossierResolver, error) {
	resolver := &directoryDossierResolver{dossiers: make(map[string]officialDossierReference)}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		ref, err := readDirectoryDossier(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		if ref.DossierRef == "" {
			return nil
		}
		resolver.dossiers[ref.DossierRef] = ref
		if ref.DossierLibelle != "" {
			resolver.byTitle = append(resolver.byTitle, ref)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk dossiers directory: %w", err)
	}
	sort.SliceStable(resolver.byTitle, func(i, j int) bool {
		return len(normalizeDossierTitle(resolver.byTitle[i].DossierLibelle)) > len(normalizeDossierTitle(resolver.byTitle[j].DossierLibelle))
	})

	return resolver, nil
}

func (r *directoryDossierResolver) Resolve(dossierRef, title, object, sortCode string) officialDossierReference {
	if dossierRef != "" {
		if ref, ok := r.dossiers[dossierRef]; ok {
			ref.Text = chooseDossierTextFromRefs(ref.Texts, title, sortCode)
			return ref
		}
	}

	text := normalizeDossierTitle(title + " " + object)
	for _, candidate := range r.byTitle {
		dossierTitle := normalizeDossierTitle(candidate.DossierLibelle)
		if len(dossierTitle) >= 20 && strings.Contains(text, dossierTitle) {
			candidate.Text = chooseDossierTextFromRefs(candidate.Texts, title, sortCode)
			return candidate
		}
	}

	return officialDossierReference{}
}

func readDirectoryDossier(path string) (officialDossierReference, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return officialDossierReference{}, err
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return officialDossierReference{}, err
	}
	dossier := objectAt(root, "dossierParlementaire")
	uid := stringAt(dossier, "uid")
	if uid == "" {
		return officialDossierReference{}, nil
	}

	refs := collectOfficialTextReferences(dossier)
	return officialDossierReference{
		DossierRef:     uid,
		DossierLibelle: stringAt(dossier, "titreDossier", "titre"),
		Text:           chooseDossierTextFromRefs(refs, "", ""),
		Texts:          refs,
	}, nil
}

func collectOfficialTextReferences(value any) []officialTextReference {
	seen := make(map[string]bool)
	var refs []officialTextReference
	var visit func(any)
	visit = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			for _, value := range typed {
				visit(value)
			}
		case []any:
			for _, value := range typed {
				visit(value)
			}
		default:
			text, ok := stringValue(typed)
			if !ok {
				return
			}
			ref := officialTextReferenceFromUID(text)
			if ref.URL == "" || seen[ref.URL] {
				return
			}
			seen[ref.URL] = true
			refs = append(refs, ref)
		}
	}
	visit(value)
	sortOfficialTextReferences(refs)
	return refs
}

func sortOfficialTextReferences(refs []officialTextReference) {
	sort.SliceStable(refs, func(i, j int) bool {
		leftRank := textReferenceKindRank(refs[i].Kind)
		rightRank := textReferenceKindRank(refs[j].Kind)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		leftNum := numericTextNum(refs[i].Num)
		rightNum := numericTextNum(refs[j].Num)
		if leftNum != rightNum {
			return leftNum < rightNum
		}
		return refs[i].URL < refs[j].URL
	})
}

func textReferenceKindRank(kind string) int {
	switch kind {
	case "projet-loi", "proposition-loi":
		return 0
	case "texte-adopte-commission":
		return 1
	case "texte-adopte-seance":
		return 2
	default:
		return 3
	}
}

func chooseDossierTextFromRefs(refs []officialTextReference, title, sortCode string) officialTextReference {
	if len(refs) == 0 {
		return officialTextReference{}
	}
	if isCMPVote(title) {
		if ref := latestTextReferenceOfKind(refs, "texte-adopte-commission"); ref.URL != "" {
			return ref
		}
	}
	if sortCode == "adopté" {
		if ref := latestTextReferenceOfKind(refs, "texte-adopte-seance"); ref.URL != "" {
			return ref
		}
	}
	if ref := firstOriginalTextReference(refs); ref.URL != "" {
		return ref
	}
	return refs[0]
}

func latestTextReferenceOfKind(refs []officialTextReference, kind string) officialTextReference {
	var selected officialTextReference
	for _, ref := range refs {
		if ref.Kind != kind {
			continue
		}
		if selected.URL == "" || numericTextNum(ref.Num) > numericTextNum(selected.Num) {
			selected = ref
		}
	}
	return selected
}

func firstOriginalTextReference(refs []officialTextReference) officialTextReference {
	for _, ref := range refs {
		if ref.Kind == "projet-loi" || ref.Kind == "proposition-loi" {
			return ref
		}
	}
	return officialTextReference{}
}

func officialTextReferenceFromUID(uid string) officialTextReference {
	if match := texteCommissionRE.FindStringSubmatch(uid); len(match) == 4 {
		return officialTextReferenceFromParts(match[2], match[3], "texte-adopte-commission", "b")
	}
	if match := texteAdopteRE.FindStringSubmatch(uid); len(match) == 4 {
		return officialTextReferenceFromParts(match[2], match[3], "texte-adopte-seance", "t")
	}
	if match := texteOriginalRE.FindStringSubmatch(uid); len(match) == 4 {
		kind := "projet-loi"
		if match[1] == "PION" {
			kind = "proposition-loi"
		}
		return officialTextReferenceFromParts(match[2], match[3], kind, "b")
	}
	return officialTextReference{}
}

func officialTextReferenceFromParts(legislature, num, kind, prefix string) officialTextReference {
	if legislature == "" || num == "" || kind == "" || prefix == "" {
		return officialTextReference{}
	}
	pathNum := leftPad(num, 4)
	base := "https://www.assemblee-nationale.fr/dyn/" + legislature + "/textes/l" + legislature + prefix + pathNum + "_" + kind
	return officialTextReference{Num: num, Kind: kind, URL: base, PDFURL: base + ".pdf"}
}

func isCMPVote(title string) bool {
	normalized := normalizeDossierTitle(title)
	return strings.Contains(normalized, "commission mixte paritaire") || strings.Contains(normalized, "texte de la commission")
}

func normalizeDossierTitle(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	lastSpace := false
	for _, r := range value {
		r = normalizeASCII(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			builder.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(builder.String())
}

func normalizeASCII(r rune) rune {
	switch r {
	case 'à', 'â', 'ä':
		return 'a'
	case 'ç':
		return 'c'
	case 'é', 'è', 'ê', 'ë':
		return 'e'
	case 'î', 'ï':
		return 'i'
	case 'ô', 'ö':
		return 'o'
	case 'ù', 'û', 'ü':
		return 'u'
	case 'ÿ':
		return 'y'
	default:
		return r
	}
}

func numericTextNum(value string) int {
	num, _ := strconv.Atoi(value)
	return num
}

func leftPad(value string, length int) string {
	for len(value) < length {
		value = "0" + value
	}
	return value
}
