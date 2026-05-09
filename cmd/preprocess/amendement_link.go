package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const publicSessionOrganeRef = "PO838901"

type officialAmendementReference struct {
	TextNum       string
	Organe        string
	OrganeRef     string
	AmendementNum string
	SeanceRef     string
	URL           string
}

type amendementLinkResolver interface {
	Resolve(legislature int, dossierRef, amendementNum, organeRef, seanceRef string) (officialAmendementReference, error)
}

type directoryAmendementResolver struct {
	refs map[string][]officialAmendementReference
}

var (
	texteLegislatifNumRE  = regexp.MustCompile(`B(?:TC)?0*(\d+)`)
	amendementOrganeRefRE = regexp.MustCompile(`L\d+(PO\d+)B`)
	amendementLookupNumRE = regexp.MustCompile(`(?i)(\d+[[:alpha:]]?)`)
)

func newDirectoryAmendementResolver(root string) (*directoryAmendementResolver, error) {
	resolver := &directoryAmendementResolver{refs: make(map[string][]officialAmendementReference)}
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
		ref, dossierRef, err := readDirectoryAmendement(path, rel)
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		if dossierRef == "" || ref.TextNum == "" || ref.Organe == "" || ref.AmendementNum == "" {
			return nil
		}
		key := amendementReferenceKey(dossierRef, ref.AmendementNum)
		resolver.refs[key] = append(resolver.refs[key], ref)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk amendements directory: %w", err)
	}
	return resolver, nil
}

func (r *directoryAmendementResolver) Resolve(legislature int, dossierRef, amendementNum, organeRef, seanceRef string) (officialAmendementReference, error) {
	if legislature <= 0 || dossierRef == "" || amendementNum == "" {
		return officialAmendementReference{}, nil
	}
	candidates := r.refs[amendementReferenceKey(dossierRef, amendementNum)]
	return chooseOfficialAmendementReference(candidates, amendementNum, organeRef, seanceRef), nil
}

func readDirectoryAmendement(path, relativePath string) (officialAmendementReference, string, error) {
	parts := strings.Split(filepath.ToSlash(relativePath), "/")
	if len(parts) < 4 {
		return officialAmendementReference{}, "", nil
	}
	dossierRef := parts[1]

	data, err := os.ReadFile(path)
	if err != nil {
		return officialAmendementReference{}, "", err
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return officialAmendementReference{}, "", err
	}
	amendement := objectAt(root, "amendement")
	legislature, _ := intAt(amendement, "legislature")
	uid := stringAt(amendement, "uid")
	ref := officialAmendementReference{
		TextNum:       textNumFromTexteLegislatifRef(stringAt(amendement, "texteLegislatifRef")),
		Organe:        stringAt(amendement, "identification", "prefixeOrganeExamen"),
		OrganeRef:     organeRefFromAmendementUID(uid),
		AmendementNum: stringAt(amendement, "identification", "numeroLong"),
		SeanceRef:     stringAt(amendement, "seanceDiscussionRef"),
	}
	ref.URL = officialAmendementPDFURL(legislature, ref)
	return ref, dossierRef, nil
}

func chooseOfficialAmendementReference(candidates []officialAmendementReference, amendementNum, organeRef, seanceRef string) officialAmendementReference {
	if seanceRef != "" {
		for _, candidate := range candidates {
			if candidate.SeanceRef == seanceRef && candidate.OrganeRef == organeRef && amendmentNumbersMatch(candidate.AmendementNum, amendementNum) {
				return candidate
			}
		}
	}
	for _, candidate := range candidates {
		if candidate.OrganeRef == organeRef && amendmentNumbersMatch(candidate.AmendementNum, amendementNum) {
			return candidate
		}
	}
	if organeRef == publicSessionOrganeRef {
		for _, candidate := range candidates {
			if candidate.Organe == "AN" && amendmentNumbersMatch(candidate.AmendementNum, amendementNum) {
				return candidate
			}
		}
	}
	for _, candidate := range candidates {
		if amendmentNumbersMatch(candidate.AmendementNum, amendementNum) {
			return candidate
		}
	}
	return officialAmendementReference{}
}

func officialAmendementPDFURL(legislature int, ref officialAmendementReference) string {
	if legislature <= 0 || ref.TextNum == "" || ref.Organe == "" || ref.AmendementNum == "" {
		return ""
	}
	return "https://www.assemblee-nationale.fr/dyn/" + strconv.Itoa(legislature) + "/amendements/" + ref.TextNum + "/" + ref.Organe + "/" + ref.AmendementNum + ".pdf"
}

func textNumFromTexteLegislatifRef(ref string) string {
	match := texteLegislatifNumRE.FindStringSubmatch(ref)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func organeRefFromAmendementUID(uid string) string {
	match := amendementOrganeRefRE.FindStringSubmatch(uid)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func amendementReferenceKey(dossierRef, amendementNum string) string {
	return dossierRef + "|" + normalizeAmendementLookupNum(amendementNum)
}

func amendmentNumbersMatch(left, right string) bool {
	return normalizeAmendementLookupNum(left) == normalizeAmendementLookupNum(right)
}

func normalizeAmendementLookupNum(value string) string {
	value = strings.ToUpper(strings.Join(strings.Fields(value), ""))
	match := amendementLookupNumRE.FindStringSubmatch(value)
	if len(match) < 2 {
		return value
	}
	return match[1]
}
